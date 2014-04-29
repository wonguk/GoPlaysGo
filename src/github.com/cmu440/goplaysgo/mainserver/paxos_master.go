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

type paxosMaster struct {
	n         int
	maxCmdNum int

	commands map[int]paxosrpc.Command
	cmdDone  map[int]chan struct{} // closed chan => Command has been run

	commandChan chan paxosrpc.Command // Requests to initialize an AI
	prepareChan chan prepRequest      // Request to check n
	acceptChan  chan acceptRequest    // Request to Accept value
	commitChan  chan paxosrpc.Command // Request to commit

	servers []node
}

type paxosHandler struct {
	n       int
	command paxosrpc.Command
}

type commitHandler struct {
	done    chan struct{}
	command paxosrpc.Command
}

func (pm *paxosMaster) startPaxosMaster(statsChan chan paxosrpc.Command) {
	// Mark Initial command as done (so first command can run)
	LOGV.Println("PaxosMaster:", "Starting Paxos Master!")
	pm.cmdDone[0] = make(chan struct{})
	close(pm.cmdDone[0])

	for {
		select {
		case c := <-pm.commandChan:
			LOGV.Println("PaxosMaster:", "Recieved Command", c.CommandNumber, c.Type)
			pm.n++

			if c.CommandNumber == -1 {
				pm.maxCmdNum++
				LOGV.Println("PaxosMaster:", "Giving CmdNum", pm.maxCmdNum)
				c.CommandNumber = pm.maxCmdNum
			}

			ph := new(paxosHandler)
			ph.n = pm.n
			ph.command = c

			go ph.startHandler(pm.commandChan, pm.commitChan, pm.servers)

		case p := <-pm.prepareChan:
			LOGV.Println("PaxosMaster:", "Recieved Prepare", p.n)
			if p.n < pm.n {
				LOGV.Println("PaxosMaster:", "Rejecting Due to Low n", p.n, pm.n)
				p.reply.Status = paxosrpc.Reject
				p.reply.N = pm.n
			} else if last, ok := pm.commands[p.cmdNum]; ok {
				LOGV.Println("PaxosMaster:", "Rejecting Due to Past Commit", p.cmdNum)
				p.reply.Status = paxosrpc.Reject
				p.reply.Command = last
				p.reply.MaxCmdNum = pm.maxCmdNum
				pm.n = p.n
			} else {
				LOGV.Println("PaxosMaster:", "Accepting Prepare", p.n)
				p.reply.Status = paxosrpc.OK
				pm.n = p.n
			}

			close(p.retChan)

		case a := <-pm.acceptChan:
			LOGV.Println("PaxosMAster:", "Recieved Accept", a.n)
			if a.n != pm.n {
				LOGV.Println("PaxosMaster:", "Rejecting due to Wrong n", a.n, pm.n)
				a.reply.Status = paxosrpc.Reject
			} else {
				LOGV.Println("PaxosMaster:", "Accepting Accept", a.n, a.command)
				a.reply.Status = paxosrpc.OK
				LOGV.Println("PaxosMaster:", "ZXCZXCZXAccepting Accept", a.n, a.command)
				pm.commitChan <- a.command
				LOGV.Println("PaxosMaster:", "ASDASDAccepting Accept", a.n, a.command)
			}

			LOGV.Println("PaxosMaster:", "QWEQWEAccepting Accept", a.n, a.command)
			close(a.retChan)

		case cmd := <-pm.commitChan:
			LOGV.Println("PaxosMaster:", "Recieved Commit", cmd)
			if cmd.CommandNumber > pm.maxCmdNum {
				pm.maxCmdNum = cmd.CommandNumber
			}

			if _, ok := pm.commands[cmd.CommandNumber]; !ok {
				LOGV.Println("PaxosMaster:", "Updaing Command!")
				//Check Prev Commands filled up

				pm.commands[cmd.CommandNumber] = cmd
				done, ok := pm.cmdDone[cmd.CommandNumber]

				if !ok {
					done = make(chan struct{})
					pm.cmdDone[cmd.CommandNumber] = done
				}

				ch := commitHandler{done, cmd}

				prevDone, ok := pm.cmdDone[cmd.CommandNumber-1]

				if !ok {
					prevDone = make(chan struct{})
					pm.cmdDone[cmd.CommandNumber-1] = prevDone

					nopCmd := paxosrpc.Command{
						Type:          paxosrpc.NOP,
						CommandNumber: cmd.CommandNumber - 1,
					}
					pm.commandChan <- nopCmd
				}

				go ch.startHandler(prevDone, statsChan)
			}
		}
	}
}

// commitHandler waits until previous commands have been run.
// Once the previous commands have been run, it runs the relevant command
func (ch *commitHandler) startHandler(done chan struct{}, commitChan chan paxosrpc.Command) {
	<-done

	commitChan <- ch.command
	close(ch.done)
}

// paxosHandler basically runs the Paxos protocol(?) on the given command
func (ph *paxosHandler) startHandler(cmdChan chan paxosrpc.Command,
	cmtChan chan paxosrpc.Command, servers []node) {
	//Prepare Phase
	LOGV.Println("PaxosHandler:", "Sending Prepare Messages")
	if !ph.prepare(cmdChan, cmtChan, servers) {
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
func (ph *paxosHandler) prepare(cmdChan chan paxosrpc.Command,
	cmtChan chan paxosrpc.Command, servers []node) bool {
	numPrepare := 0
	prepareChan := make(chan *rpc.Call, len(servers))

	pArgs := paxosrpc.PrepareArgs{ph.n, ph.command.CommandNumber}

	for _, s := range servers {
		var pReply paxosrpc.PrepareReply
		call := s.client.Go("PaxosServer.Prepare", &pArgs, &pReply, nil)

		go rpcHandler(call, prepareChan)
	}

	for i := 0; i < len(servers); i++ {
		call := <-prepareChan

		if call.Error != nil {
			continue
		}

		reply := (call.Reply).(*paxosrpc.PrepareReply)

		switch reply.Status {
		case paxosrpc.OK:
			numPrepare++
			if numPrepare == len(servers)/2 {
				break
			}

		case paxosrpc.Reject:
			if reply.Command.CommandNumber == ph.command.CommandNumber {
				cmtChan <- reply.Command
			}
			if reply.MaxCmdNum > ph.command.CommandNumber {
				ph.command.CommandNumber = reply.MaxCmdNum + 1
			} else {
				ph.command.CommandNumber++
			}
			cmdChan <- ph.command
			return false

		default:
			continue
		}
	}

	// If not enough OKs, resend command to master
	if numPrepare < len(servers)/2 {
		ph.command.CommandNumber++
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
		ph.command.CommandNumber++
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
		var cReply paxosrpc.CommitReply
		s.client.Go("PaxosServer", &cArgs, &cReply, nil)
	}
}

// When the rpc call is done, returns it to the given channel
func rpcHandler(call *rpc.Call, retChan chan *rpc.Call) {
	done := <-call.Done
	retChan <- done
}
