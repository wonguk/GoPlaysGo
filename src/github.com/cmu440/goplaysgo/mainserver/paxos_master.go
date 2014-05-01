package mainserver

import (
	"net/rpc"

	"github.com/cmu440/goplaysgo/rpc/paxosrpc"
)

type prepRequest struct {
	n      int
	cmdNum int
	reply  *paxosrpc.PrepareReply

	retChan chan struct{}
}

type acceptRequest struct {
	n       int
	command paxosrpc.Command
	reply   *paxosrpc.AcceptReply

	retChan chan struct{}
}

type maxRequest struct {
	retChan chan maxReply
}

type maxReply struct {
	done   chan bool
	cmdNum int
}

type doneRequest struct {
	cmdNum  int
	retChan chan doneReply
}

type doneReply struct {
	cmdNum int
	done   chan bool
}

type cmdRequest struct {
	retChan chan []paxosrpc.Command
}

type paxosMaster struct {
	myn       int
	n         map[int]int
	maxCmdNum int

	uncommited  map[int]paxosrpc.Command
	uncommitedN map[int]int
	commands    map[int]paxosrpc.Command
	cmdDone     map[int]chan bool // closed chan => Command has been run

	nChan       chan cn
	commandChan chan paxosrpc.Command // Requests to initialize an AI
	prepareChan chan prepRequest      // Request to check n
	acceptChan  chan acceptRequest    // Request to Accept value
	commitChan  chan paxosrpc.Command // Request to commit
	getCmdChan  chan cmdRequest
	doneChan    chan doneRequest
	maxChan     chan maxRequest
	serverChan  chan []node

	servers []node
}

type cn struct {
	cmdNum int
	n      int
}

type paxosHandler struct {
	n       int
	command paxosrpc.Command
}

type commitHandler struct {
	done    chan bool
	command paxosrpc.Command
}

// Paxos Master
// The Paxos Master is responsible for communicating with the other Main Servers
// in the Paxos ring and commiting values that have been requested locally and
// from other servers in the ring.
// The Paxos Master keeps track of all the commands that have been commited until
// now and makes sure that for each command, the previous command has been run.

func (pm *paxosMaster) startPaxosMaster(statsChan chan paxosrpc.Command) {
	// Mark Initial command as done (so first command can run)
	LOGV.Println("PaxosMaster:", "Starting Paxos Master!")
	pm.commands[0] = paxosrpc.Command{
		CommandNumber: 0,
		Type:          paxosrpc.NOP,
	}
	pm.cmdDone[0] = make(chan bool, 1)
	close(pm.cmdDone[0])

	for {
		select {
		case cn := <-pm.nChan:
			if cn.n > pm.n[cn.cmdNum] {
				pm.n[cn.cmdNum] = cn.n
			}

		case c := <-pm.commandChan:
			// Requesting a Paxos Commit
			LOGV.Println("PaxosMaster:", "Recieved Command", c.CommandNumber, c.Type)

			if c.CommandNumber == -1 {
				pm.maxCmdNum++
				LOGV.Println("PaxosMaster:", "Giving CmdNum", pm.maxCmdNum)
				c.CommandNumber = pm.maxCmdNum
			}

			ph := new(paxosHandler)
			n, ok := pm.n[c.CommandNumber]
			if !ok {
				pm.n[c.CommandNumber] = 0
				n = 0
			}
			pm.n[c.CommandNumber] = n + 1
			ph.n = n + 1
			ph.command = c

			go ph.startHandler(pm.nChan, pm.commandChan, pm.commitChan, pm.servers)

		case p := <-pm.prepareChan:
			// Prepare Request
			LOGV.Println("PaxosMaster:", "Recieved Prepare", p.n)

			n, ok := pm.n[p.cmdNum]

			if !ok {
				pm.n[p.cmdNum] = 0
				n = 0
			}

			if last, ok := pm.commands[p.cmdNum]; ok {
				LOGV.Println("PaxosMaster:", "Rejecting Due to Past Commit", p.cmdNum, last)
				p.reply.Status = paxosrpc.Reject
				p.reply.Command = last
				p.reply.MaxCmdNum = pm.maxCmdNum
				if p.n > n {
					pm.n[p.cmdNum] = p.n
				}
			} else if p.n < n {
				LOGV.Println("PaxosMaster:", "Rejecting Due to Low n", p.n, n)
				p.reply.Status = paxosrpc.Reject
				p.reply.N = n
			} else if uncommited, ok := pm.uncommited[p.cmdNum]; ok {
				LOGV.Println("PaxosMaster:", "Accepting, but has uncommited val")

				p.reply.Status = paxosrpc.OK
				p.reply.Command = uncommited
				p.reply.N = pm.uncommitedN[p.cmdNum]
				p.reply.MaxCmdNum = pm.maxCmdNum
				pm.n[p.cmdNum] = p.n
			} else {
				LOGV.Println("PaxosMaster:", "Accepting Prepare", p.n)
				p.reply.Status = paxosrpc.OK
				pm.n[p.cmdNum] = p.n
			}

			close(p.retChan)

		case a := <-pm.acceptChan:
			// Accept Request
			LOGV.Println("PaxosMAster:", "Recieved Accept", a.n)
			n, ok := pm.n[a.command.CommandNumber]

			if !ok {
				pm.n[a.command.CommandNumber] = 0
				n = 0
			}

			if a.n < n {
				LOGV.Println("PaxosMaster:", "Rejecting due to Wrong n", a.n, n)
				a.reply.Status = paxosrpc.Reject
			} else {
				LOGV.Println("PaxosMaster:", "Accepting Accept", a.n, a.command)
				a.reply.Status = paxosrpc.OK
				pm.uncommited[a.command.CommandNumber] = a.command
				pm.uncommitedN[a.command.CommandNumber] = a.n
				pm.n[a.command.CommandNumber] = a.n
			}

			close(a.retChan)

		case cmd := <-pm.commitChan:
			// Commit Request
			LOGS.Println("PaxosMaster:", "Recieved Commit", cmd)
			if cmd.CommandNumber > pm.maxCmdNum {
				pm.maxCmdNum = cmd.CommandNumber
			}

			if _, ok := pm.commands[cmd.CommandNumber]; !ok {
				LOGS.Println("PaxosMaster:", "Updaing Command!")

				// Update Commands
				pm.commands[cmd.CommandNumber] = cmd
				done, ok := pm.cmdDone[cmd.CommandNumber]

				if !ok {
					LOGS.Println("PaxosMaster:", "Initializing Done Chan for", cmd.CommandNumber)
					done = make(chan bool, 1)
					pm.cmdDone[cmd.CommandNumber] = done
				}

				ch := commitHandler{done, cmd}

				//Check Prev Commands filled up
				prevDone, ok := pm.cmdDone[cmd.CommandNumber-1]

				if !ok {
					LOGS.Println("PaxosMaster:", "Initializing Prev Step", cmd.CommandNumber-1)
					prevDone = make(chan bool, 1)
					pm.cmdDone[cmd.CommandNumber-1] = prevDone

					nopCmd := paxosrpc.Command{
						Type:          paxosrpc.NOP,
						CommandNumber: cmd.CommandNumber - 1,
					}
					pm.commandChan <- nopCmd
				}

				go ch.startHandler(prevDone, statsChan)
			}

		case req := <-pm.doneChan:
			LOGS.Println("PaxosMaster:", "Setup Done Chan", req.cmdNum)
			done, ok := pm.cmdDone[req.cmdNum]

			if !ok {
				LOGS.Println("PaxosMaster:", "Making new Done Chan", req.cmdNum)
				done = make(chan bool)
				pm.cmdDone[req.cmdNum] = done
			}

			reply := doneReply{req.cmdNum, done}
			req.retChan <- reply

		case req := <-pm.maxChan:
			done, ok := pm.cmdDone[pm.maxCmdNum]

			if !ok {
				done := make(chan bool)
				pm.cmdDone[pm.maxCmdNum] = done
			}

			reply := maxReply{done, pm.maxCmdNum}
			req.retChan <- reply

		case req := <-pm.getCmdChan:
			cmds := make([]paxosrpc.Command, len(pm.commands))

			for i, c := range pm.commands {
				cmds[i] = c
			}
			req.retChan <- cmds

		case servers := <-pm.serverChan:
			pm.servers = servers

		}
	}
}

// commitHandler waits until previous commands have been run.
// Once the previous commands have been run, it runs the relevant command
func (ch *commitHandler) startHandler(done chan bool, commitChan chan paxosrpc.Command) {
	LOGV.Println("CommitHandler:", "Waiting for prevTask to be done!", ch.command.CommandNumber)
	<-done

	LOGS.Println("CommitHandler:", "PrevTask for", ch.command.CommandNumber, "is Done!")
	commitChan <- ch.command
	close(ch.done)
}

// paxosHandler basically runs the Paxos protocol(?) on the given command
func (ph *paxosHandler) startHandler(nChan chan cn, cmdChan chan paxosrpc.Command,
	cmtChan chan paxosrpc.Command, servers []node) {
	//Prepare Phase
	LOGV.Println("PaxosHandler:", "Sending Prepare Messages")
	if !ph.prepare(nChan, cmdChan, cmtChan, servers) {
		LOGE.Println("PaxosHandler:", "Failed To Prepare Message", ph.command)
		return
	}

	//Accept Phase
	LOGV.Println("PaxosHandler:", "Sending Accept Messages")
	if !ph.accept(cmdChan, servers) {
		LOGE.Println("PaxosHandler:", "Failed To Accept Message", ph.command)
		return
	}

	//Commit Phase
	LOGV.Println("PaxosHandler:", "Sending Commit Messages")
	ph.commit(cmtChan, servers)
}

// The Prepare phase in paxos
func (ph *paxosHandler) prepare(nChan chan cn, cmdChan chan paxosrpc.Command,
	cmtChan chan paxosrpc.Command, servers []node) bool {
	numPrepare := 0
	prepareChan := make(chan *rpc.Call, len(servers))

	pArgs := paxosrpc.PrepareArgs{ph.n, ph.command.CommandNumber}

	for _, s := range servers {
		if s.client == nil {
			continue
		}
		var pReply paxosrpc.PrepareReply
		call := s.client.Go("PaxosServer.Prepare", &pArgs, &pReply, nil)

		go rpcHandler(call, prepareChan)
	}

	maxCmd := 0
	for i := 0; i < len(servers); i++ {
		call := <-prepareChan

		if call.Error != nil {
			continue
		}

		reply := (call.Reply).(*paxosrpc.PrepareReply)

		switch reply.Status {
		case paxosrpc.OK:
			numPrepare++

			if reply.N > ph.n {
				if ph.command.Type == paxosrpc.NOP {
					ph.command = reply.Command
				}
				ph.n = reply.N
			}

			if numPrepare == len(servers)/2 {
				break
			}

		case paxosrpc.Reject:
			nChan <- cn{reply.N, reply.Command.CommandNumber}

			// If there was an already commited value at CommandNumber, Give up
			if reply.Command.CommandNumber == ph.command.CommandNumber {
				cmtChan <- reply.Command
				if ph.command.Type != paxosrpc.NOP {
					ph.command.CommandNumber = -1
					cmdChan <- ph.command
				}
				return false
			}
			if reply.MaxCmdNum > maxCmd {
				maxCmd = reply.MaxCmdNum
			}

		default:
			continue
		}
	}

	// If not enough OKs, resend command to master
	if numPrepare < len(servers)/2 {
		if ph.command.Type != paxosrpc.NOP {
			ph.command.CommandNumber = -1
		}
		cmdChan <- ph.command
		return false
	}

	return true
}

// The Accept phase in paxos
func (ph *paxosHandler) accept(cmdChan chan paxosrpc.Command, servers []node) bool {
	numAccept := 0
	acceptChan := make(chan *rpc.Call, len(servers))

	aArgs := paxosrpc.AcceptArgs{ph.n, ph.command}
	for _, s := range servers {
		if s.client == nil {
			continue
		}
		var aReply paxosrpc.AcceptReply
		call := s.client.Go("PaxosServer.Accept", &aArgs, &aReply, nil)

		go rpcHandler(call, acceptChan)
	}

	for i := 0; i < len(servers); i++ {
		call := <-acceptChan

		if call.Error != nil {
			continue
		}

		reply := (call.Reply).(*paxosrpc.AcceptReply)

		switch reply.Status {
		case paxosrpc.OK:
			numAccept++
			if numAccept == len(servers)/2 {
				break
			}
		default:
			continue
		}
	}

	// If not enough OKs, resend command to master
	if numAccept < len(servers)/2 {
		if ph.command.Type != paxosrpc.NOP {
			ph.command.CommandNumber = -1
		}
		cmdChan <- ph.command
		return false
	}

	return true
}

// The Commit phase is paxos
func (ph *paxosHandler) commit(cmtChan chan paxosrpc.Command, servers []node) {
	cmtChan <- ph.command

	cArgs := paxosrpc.CommitArgs{ph.n, ph.command}
	for _, s := range servers {
		if s.client == nil {
			continue
		}
		var cReply paxosrpc.CommitReply
		s.client.Go("PaxosServer.Commit", &cArgs, &cReply, nil)
	}
}

// When the rpc call is done, returns it to the given channel
func rpcHandler(call *rpc.Call, retChan chan *rpc.Call) {
	done := <-call.Done
	retChan <- done
}
