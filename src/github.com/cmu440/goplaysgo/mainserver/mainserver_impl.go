package mainserver

import (
	"container/list"
	"errors"
	//"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cmu440/goplaysgo/rpc/mainrpc"
	"github.com/cmu440/goplaysgo/rpc/paxosrpc"
)

// Error Log
var LOGE = log.New(os.Stdout, "ERROR [MainServer] ",
	log.Lmicroseconds|log.Lshortfile)

// Verbose Log
var LOGV = log.New(os.Stdout, "VERBOSE [MainServer] ", log.Lmicroseconds|log.Lshortfile)
var LOGS = log.New(os.Stdout, "VERBOSE [MainServer] ", log.Lmicroseconds|log.Lshortfile)

type isReady struct {
	sync.Mutex
	ready bool
}

type node struct {
	hostport string
	client   *rpc.Client
}

type mainServer struct {
	master      bool
	replacement bool
	masterLock  sync.Mutex
	readyLock   sync.Mutex
	masterAddr  string
	numNodes    int
	port        int
	hostport    string
	servers     []node

	ready       chan struct{}
	masterReady chan struct{}
	isReady     isReady
	quiese      isReady

	aiMaster    *aiMaster
	statsMaster *statsMaster
	paxosMaster *paxosMaster
}

// NewMainServer returns a mainserver that manages the different AIs and
// their stats
func NewMainServer(masterServerHostPort string, numNodes, port int, isReplacement bool) (MainServer, error) {
	ms := new(mainServer)

	if isReplacement {
		LOGV.Println("Starting Replacement", port)
		ms.replacement = true
	} else if masterServerHostPort == "" {
		LOGV.Println("Starting Master...:", port)
		ms.master = true
	} else {
		LOGV.Println("Starting Slave...:", port)
		ms.master = false
	}

	ms.hostport = "localhost:" + strconv.Itoa(port)

	ms.masterLock = sync.Mutex{}
	ms.readyLock = sync.Mutex{}
	ms.masterAddr = masterServerHostPort
	ms.numNodes = numNodes
	ms.port = port

	ms.servers = []node{node{hostport: ms.hostport}}

	ms.ready = make(chan struct{})
	ms.masterReady = make(chan struct{})

	rpc.RegisterName("MainServer", ms)
	rpc.RegisterName("PaxosServer", ms)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", ":"+strconv.Itoa(port))
	if e != nil {
		LOGE.Println("Error Listening to port", port, e)
		return nil, errors.New("error listening to port" + strconv.Itoa(port))
	}
	go http.Serve(l, nil)

	if ms.replacement {
		LOGS.Println("I AM A REPLACEMENT")
		LOGS.Println("WAiting to be readyu")
		<-ms.ready
		LOGS.Println("Done!")

		ms.startMasters()

		close(ms.masterReady)

		ms.isReady.Lock()
		ms.isReady.ready = true
		ms.isReady.Unlock()

		return ms, nil
	}

	if ms.master {
		if ms.numNodes == 1 {
			LOGV.Println("Master:", "Done Initializing")
			ms.startMasters()

			ms.isReady.Lock()
			ms.isReady.ready = true
			ms.isReady.Unlock()

			return ms, nil
		}

		LOGV.Println("Master:", port, "waiting for nodes to register")
		<-ms.ready

		ms.startMasters()

		ms.isReady.Lock()
		ms.isReady.ready = true
		ms.isReady.Unlock()

		return ms, nil
	}

	LOGV.Println("Slave:", port, "Dialing master")
	client := dialHTTP(masterServerHostPort)

	args := mainrpc.RegisterArgs{ms.hostport}
	var reply mainrpc.RegisterReply

	for {
		LOGV.Println("Slave:", port, "registering to master")
		err := client.Call("MainServer.RegisterServer", &args, &reply)

		if err == nil && reply.Status == mainrpc.OK {
			LOGV.Println("Slave:", ms.port, "Registered to master!")
			ms.servers = make([]node, len(reply.Servers))

			for i, h := range reply.Servers {
				ms.servers[i].hostport = h
			}

			ms.initClients()
			ms.startMasters()

			ms.isReady.Lock()
			ms.isReady.ready = true
			ms.isReady.Unlock()

			return ms, nil
		}

		if err != nil {
			LOGE.Println("Slave:", port, "erro registereing to master", err)
		}

		LOGV.Println("Slave:", port, "sleeping for 1 second")
		LOGV.Println("Slave:", port, "Master Status", reply.Status)
		time.Sleep(time.Second)
	}

	return nil, errors.New("should have been successful")
}

// startMasters starts all the backend threads for the main servers
func (ms *mainServer) startMasters() {
	LOGS.Println("Starting Masters!")
	// Initialize PaxosMaster
	paxosMaster := new(paxosMaster)
	paxosMaster.maxCmdNum = 0
	paxosMaster.n = make(map[int]int)
	paxosMaster.uncommited = make(map[int]paxosrpc.Command)
	paxosMaster.uncommitedN = make(map[int]int)
	paxosMaster.commands = make(map[int]paxosrpc.Command)
	paxosMaster.cmdDone = make(map[int]chan bool)
	paxosMaster.nChan = make(chan cn, 1000)
	paxosMaster.commandChan = make(chan paxosrpc.Command, 1000)
	paxosMaster.prepareChan = make(chan prepRequest, 1000)
	paxosMaster.acceptChan = make(chan acceptRequest, 1000)
	paxosMaster.commitChan = make(chan paxosrpc.Command, 1000)
	paxosMaster.getCmdChan = make(chan cmdRequest, 1000)
	paxosMaster.serverChan = make(chan []node, 10)
	paxosMaster.doneChan = make(chan doneRequest, 1000)
	paxosMaster.maxChan = make(chan maxRequest, 1000)
	paxosMaster.servers = make([]node, len(ms.servers)-1)

	// Don't include self in the list of servers
	i := 0
	for _, n := range ms.servers {
		if n.hostport != ms.hostport {
			paxosMaster.servers[i] = n
			i++
		}
	}

	// Initialize StatsMaster
	statsMaster := new(statsMaster)
	statsMaster.reqChan = make(chan statsRequest, 1000)
	statsMaster.allReqChan = make(chan allStatsRequest, 1000)
	statsMaster.initChan = make(chan initRequest, 1000)
	statsMaster.resultChan = make(chan resultRequest, 1000)
	statsMaster.commitChan = make(chan paxosrpc.Command, 1000)
	statsMaster.stats = make(map[string]mainrpc.Stats)
	statsMaster.toRun = make(map[string]chan bool)
	statsMaster.toReturn = list.New()

	// Initialize AIMaster
	aiMaster := new(aiMaster)
	aiMaster.aiChan = make(chan *newAIReq, 1000)
	aiMaster.getChan = make(chan *getAIsReq, 1000)
	aiMaster.serverChan = make(chan []string, 10)
	aiMaster.aiClients = make(map[string]*aiInfo)

	// Start Masters
	go paxosMaster.startPaxosMaster(statsMaster.commitChan)
	go statsMaster.startStatsMaster(paxosMaster.commandChan, aiMaster.aiChan)
	go aiMaster.startAIMaster(statsMaster.initChan, ms.getServers()) //, statsMaster.addChan)

	ms.paxosMaster = paxosMaster
	ms.statsMaster = statsMaster
	ms.aiMaster = aiMaster

}

// RegisterServer will add the servers into the Paxos ring
// of main servers. Initially, there will be a master main server,
// which will wait for all the main servers to connect
func (ms *mainServer) RegisterServer(args *mainrpc.RegisterArgs, reply *mainrpc.RegisterReply) error {
	if !ms.master {
		LOGE.Println("RegisterServer:", "cannot register server to slave")
		return errors.New("cannot register server to a slave server")
	}

	LOGV.Println("RegisterServer:", "registering", args.Hostport)

	ms.masterLock.Lock()
	defer ms.masterLock.Unlock()

	if len(ms.servers) == ms.numNodes {
		LOGV.Println("RegisterServer:", "Registered all nodes! replying to",
			args.Hostport)
		reply.Status = mainrpc.OK
		reply.Servers = ms.getServers()

		return nil
	}

	for _, node := range ms.servers {
		if node.hostport == args.Hostport {
			reply.Status = mainrpc.NotReady
			return nil
		}
	}

	client := dialHTTP(args.Hostport)
	ms.servers = append(ms.servers, node{args.Hostport, client})

	if len(ms.servers) == ms.numNodes {
		reply.Status = mainrpc.OK
		reply.Servers = ms.getServers()

		close(ms.ready)
	} else {
		reply.Status = mainrpc.NotReady
	}

	return nil
}

// getServers returns an array of hostports of the servers
func (ms *mainServer) getServers() []string {
	servers := make([]string, len(ms.servers))

	for i, n := range ms.servers {
		servers[i] = n.hostport
	}

	return servers
}

// GetServers returns a list of all main servers that are curently
// connected in the paxos ring
func (ms *mainServer) GetServers(args *mainrpc.GetServersArgs, reply *mainrpc.GetServersReply) error {
	LOGV.Println("GetServers:")
	ms.isReady.Lock()
	if !ms.isReady.ready {
		ms.isReady.Unlock()
		reply.Status = mainrpc.NotReady
	}
	ms.isReady.Unlock()

	ms.quiese.Lock()
	if !ms.quiese.ready {
		ms.quiese.Unlock()
		reply.Status = mainrpc.NotReady
	}
	ms.quiese.Unlock()

	reply.Status = mainrpc.OK
	reply.Servers = ms.getServers()

	return nil
}

// SubmitAI takes in an AI go program and schedules them to
func (ms *mainServer) SubmitAI(args *mainrpc.SubmitAIArgs, reply *mainrpc.SubmitAIReply) error {
	LOGV.Println("SubmitAI")

	ms.isReady.Lock()
	if !ms.isReady.ready {
		ms.isReady.Unlock()

		reply.Status = mainrpc.NotReady
		return nil
	}
	ms.isReady.Unlock()

	ms.quiese.Lock()
	if !ms.isReady.ready {
		ms.quiese.Unlock()
		reply.Status = mainrpc.NotReady
		return nil
	}
	ms.quiese.Unlock()

	retChan := make(chan mainrpc.Status)
	req := &newAIReq{
		name:     args.Name,
		code:     args.Code,
		manage:   true,
		hostport: "",
		retChan:  retChan,
	}

	ms.aiMaster.aiChan <- req

	reply.Status = <-retChan

	return nil
}

func (ms *mainServer) SubmitResult(args *mainrpc.SubmitResultArgs, reply *mainrpc.SubmitResultReply) error {
	ms.isReady.Lock()
	if !ms.isReady.ready {
		ms.isReady.Unlock()
		reply.Status = mainrpc.NotReady
		return nil
	}
	ms.isReady.Unlock()

	ms.quiese.Lock()
	if !ms.isReady.ready {
		ms.quiese.Unlock()
		reply.Status = mainrpc.NotReady
		return nil
	}
	ms.quiese.Unlock()

	req := resultRequest{
		result:  args.GameResult,
		retChan: make(chan mainrpc.Status),
	}

	ms.statsMaster.resultChan <- req

	reply.Status = <-req.retChan

	return nil
}

// GetStandings returns the current standings of the different AIs
// in the server.
func (ms *mainServer) GetStandings(args *mainrpc.GetStandingsArgs, reply *mainrpc.GetStandingsReply) error {
	LOGV.Println("GetStandings:")

	ms.isReady.Lock()
	if !ms.isReady.ready {
		ms.isReady.Unlock()

		reply.Status = mainrpc.NotReady
		return nil
	}
	ms.isReady.Unlock()

	retChan := make(chan mainrpc.Standings)

	req := allStatsRequest{retChan}
	ms.statsMaster.allReqChan <- req

	reply.Standings = <-retChan
	reply.Status = mainrpc.OK

	return nil
}

// Prepare is the rpc called when a MainServer wants to propose(?) a command
func (ms *mainServer) Prepare(args *paxosrpc.PrepareArgs, reply *paxosrpc.PrepareReply) error {
	LOGV.Println("Prepare:", "Recived Prepare Request")
	ms.isReady.Lock()
	if !ms.isReady.ready {
		ms.isReady.Unlock()

		reply.Status = paxosrpc.NotReady
		return nil
	}
	ms.isReady.Unlock()

	req := prepRequest{
		n:       args.N,
		cmdNum:  args.CommandNumber,
		reply:   reply,
		retChan: make(chan struct{}),
	}

	LOGV.Println("Prepare:", "Sending Request to PaxosMaster")
	ms.paxosMaster.prepareChan <- req
	<-req.retChan
	LOGV.Println("Prepare:", "returning")

	return nil
}

// Accept is the rpc called once a majority agreed to the Prepare Call
func (ms *mainServer) Accept(args *paxosrpc.AcceptArgs, reply *paxosrpc.AcceptReply) error {
	LOGV.Println("MainServer:", "Recived Accept Request")
	ms.isReady.Lock()
	if !ms.isReady.ready {
		ms.isReady.Unlock()

		reply.Status = paxosrpc.NotReady
		return nil
	}
	ms.isReady.Unlock()

	req := acceptRequest{
		n:       args.N,
		command: args.Command,
		reply:   reply,
		retChan: make(chan struct{}),
	}

	LOGV.Println("Accept:", "Sending Request to PaxosMaster")
	ms.paxosMaster.acceptChan <- req
	<-req.retChan
	LOGV.Println("Accept:", "returning")

	return nil
}

// Commit is the rpc called once the command has been accepted by the majority
func (ms *mainServer) Commit(args *paxosrpc.CommitArgs, reply *paxosrpc.CommitReply) error {
	LOGV.Println("MainServer:", "Recived Commit Request")
	ms.isReady.Lock()
	if !ms.isReady.ready {
		ms.isReady.Unlock()

		reply.Status = paxosrpc.NotReady
		return nil
	}
	ms.isReady.Unlock()

	LOGV.Println("Commit:", "Sending Request to PaxosMaster")
	ms.paxosMaster.commitChan <- args.Command
	LOGV.Println("Commit:", "returning")

	reply.Status = paxosrpc.OK

	return nil
}

func (ms *mainServer) Quiese(args *paxosrpc.QuieseArgs, reply *paxosrpc.QuieseReply) error {
	LOGV.Println("Quiese:", "Entering Quiese")
	ms.isReady.Lock()
	if !ms.isReady.ready && args.Type != paxosrpc.CatchUp {
		ms.isReady.Unlock()
		reply.Status = paxosrpc.NotReady
		return nil
	}
	ms.isReady.Unlock()

	switch args.Type {
	case paxosrpc.Setup:
		LOGS.Println("Quiese:", "Setting Up!")
		ms.quiese.Lock()
		ms.quiese.ready = false
		ms.quiese.Unlock()

		LOGS.Println("Quiese:", "(Setup)", "Requesting Max Command Number")
		req := maxRequest{make(chan maxReply)}
		ms.paxosMaster.maxChan <- req

		maxReply := <-req.retChan
		LOGS.Println("Quiese:", "(Setup)", "Max Command Number is:", maxReply.cmdNum)

		// Wait till done
		LOGS.Println("Quiese:", "(Setup)", "Waiting for Max Command to be Done")
		<-maxReply.done
		LOGS.Println("Quiese:", "(Setup)", "Max Command Done!")

		reply.Status = paxosrpc.OK
		reply.Servers = ms.getServers()
		reply.CommandNumber = maxReply.cmdNum

	case paxosrpc.Sync:
		LOGS.Println("Quiese:", "Syncing!")

		LOGS.Println("Quiese:", "(Sync)", "Requesting Sync upto command number", args.CommandNumber)
		nopReq := paxosrpc.Command{
			CommandNumber: args.CommandNumber,
			Type:          paxosrpc.NOP,
		}

		ms.paxosMaster.commandChan <- nopReq

		LOGS.Println("Quiese:", "(Sync)", "Requesting Done Channel for cmdNum", args.CommandNumber)
		// Get Done Chan and wait till done
		req := doneRequest{args.CommandNumber, make(chan doneReply)}
		ms.paxosMaster.doneChan <- req

		done := <-req.retChan
		LOGS.Println("Quiese:", "(Sync)", "Recieved Done Channel")

		// Wait till done
		LOGS.Println("Quiese:", "(Sync)", "Waiting for cmdNum", args.CommandNumber, "to be done")
		<-done.done
		LOGS.Println("Quiese:", "(Sync)", "Done Waiting")

		reply.Status = paxosrpc.OK

	case paxosrpc.Replace:
		LOGS.Println("QUiese:", "(Replace)", "Am I a Master?", args.Master)

		LOGS.Println("QUiese:", "(Replace)", "Replacing Servers")
		i := 0

		if len(args.ToAdd) != len(args.ToReplace) {
			reply.Status = paxosrpc.Reject
			return nil
		}

		for j, n := range ms.servers {
			LOGS.Println("QUiese:", "(Replace)", "Updating", n)
			// Update Servers
			if n.hostport == args.ToReplace[i] {
				LOGS.Println("QUiese:", "(Replace)", "replacing", n.hostport, "with", args.ToAdd[i])
				n.hostport = args.ToAdd[i]
				n.client = dialHTTP(n.hostport)
				ms.servers[j] = n

			}
		}

		// Tell Masters about new Servers
		ms.paxosMaster.serverChan <- ms.servers
		ms.aiMaster.serverChan <- ms.getServers()

		// If Master, tell
		if args.Master {

			// Get Commands until now
			LOGS.Println("QUiese:", "(Replace)", "Getting Commands!")
			req := cmdRequest{make(chan []paxosrpc.Command)}
			ms.paxosMaster.getCmdChan <- req
			commands := <-req.retChan
			LOGS.Println("QUiese:", "(Replace)", "Recieved Commands:", commands)

			LOGS.Println("QUiese:", "(Replace)", "Sending CatchUP Request")
			qArgs := paxosrpc.QuieseArgs{
				Type:     paxosrpc.CatchUp,
				Commands: commands,
				Servers:  ms.getServers(),
			}

			for _, s := range args.ToAdd {
				client := dialHTTP(s)

				var qReply paxosrpc.QuieseReply
				err := client.Call("PaxosServer.Quiese", &qArgs, &qReply)

				if err != nil {
					LOGS.Println("Quiese:", "(Replace", "ERROR WHile clling catchup", err)
				}

				if qReply.Status != paxosrpc.OK {
					LOGS.Println("QWEQWEQWEQWEQWEQWEQW")
				}

				LOGS.Println("QUiese:", "(Replace)", "Done Seding CatchUP", qReply.Status)
			}
		}

		reply.Status = paxosrpc.OK

		ms.quiese.Lock()
		ms.quiese.ready = true
		ms.quiese.Unlock()

	case paxosrpc.CatchUp:
		LOGS.Println("Quiese:", "CatchUp", "Recieved CAtchup Request")
		if !ms.replacement {
			reply.Status = paxosrpc.WrongServer
			return nil
		}

		ms.isReady.Lock()
		if ms.isReady.ready {
			ms.isReady.Unlock()
			// This replacement server had already been initialized
			reply.Status = paxosrpc.WrongServer
			return nil
		}
		ms.isReady.Unlock()

		// Update Server list
		LOGS.Println("Quiese:", "CathUp", "Updaing Server List", args.Servers)
		ms.servers = make([]node, len(args.Servers))

		for i, h := range args.Servers {
			ms.servers[i].hostport = h
		}
		ms.initClients()

		reply.Status = paxosrpc.OK
		// Close ready
		LOGS.Println("Quiese:", "CathUp", "Closing Ready to start up Masters")
		close(ms.ready)

		LOGS.Println("Wait For Master To End")
		<-ms.masterReady

		// Send Commands to commit chan of the Paxos Master
		LOGS.Println("Quiese:", "CathUp", "Commiting Commands")
		for _, c := range args.Commands {
			ms.paxosMaster.commitChan <- c
		}

	}

	return nil
}

func (ms *mainServer) initClients() {
	LOGV.Println("InitClients:", "initializing clients", ms.servers)
	for i, n := range ms.servers {
		if n.hostport != ms.hostport {
			ms.servers[i].client = dialHTTP(n.hostport)
		}
	}
	LOGV.Println("InitClients:", "initializing clients", ms.servers)
}

func dialHTTP(hostport string) *rpc.Client {
	LOGV.Println("Dialing", hostport)
	client, err := rpc.DialHTTP("tcp", hostport)

	for err != nil {
		LOGV.Println("Dialing", hostport, err)
		time.Sleep(time.Second)
		client, err = rpc.DialHTTP("tcp", hostport)
	}

	return client
}
