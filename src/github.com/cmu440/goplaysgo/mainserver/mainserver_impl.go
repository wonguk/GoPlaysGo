package mainserver

import (
	"container/list"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"
	"time"

	"github.com/cmu440/goplaysgo/rpc/mainrpc"
	"github.com/cmu440/goplaysgo/rpc/paxosrpc"
)

// Error Log
var LOGE = log.New(ioutil.Discard, "ERROR [MainServer] ",
	log.Lmicroseconds|log.Lshortfile)

// Verbose Log
var LOGV = log.New(ioutil.Discard, "VERBOSE [MainServer] ", log.Lmicroseconds|log.Lshortfile)

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

	ready   chan struct{}
	isReady isReady
	quiese  isReady

	aiMaster    *aiMaster
	statsMaster *statsMaster
	paxosMaster *paxosMaster
}

// NewMainServer returns a mainserver that manages the different AIs and
// their stats
func NewMainServer(masterServerHostPort string, numNodes, port int, isReplacement bool) (MainServer, error) {
	ms := new(mainServer)

	if masterServerHostPort == "" {
		LOGV.Println("Starting Master...:", port)
		ms.master = true
	} else {
		LOGV.Println("Starting Slave...:", port)
		ms.master = false
	}

	//TODO Possibly change later to test distributed impl
	ms.hostport = "localhost:" + strconv.Itoa(port)

	ms.masterLock = sync.Mutex{}
	ms.readyLock = sync.Mutex{}
	ms.masterAddr = masterServerHostPort
	ms.numNodes = numNodes
	ms.port = port

	ms.servers = []node{node{hostport: ms.hostport}}

	ms.ready = make(chan struct{})

	rpc.RegisterName("MainServer", ms)
	rpc.RegisterName("PaxosServer", ms)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", ":"+strconv.Itoa(port))
	if e != nil {
		LOGE.Println("Error Listening to port", port, e)
		return nil, errors.New("error listening to port" + strconv.Itoa(port))
	}
	go http.Serve(l, nil)

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

	if isReplacement {
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
	//statsMaster.addChan = make(chan mainrpc.GameResult, 1000)
	statsMaster.resultChan = make(chan resultRequest, 1000)
	statsMaster.commitChan = make(chan paxosrpc.Command, 1000)
	statsMaster.stats = make(map[string]mainrpc.Stats)
	statsMaster.toRun = make(map[string]chan bool)
	statsMaster.toReturn = list.New()

	// Initialize AIMaster
	aiMaster := new(aiMaster)
	aiMaster.aiChan = make(chan *newAIReq, 1000)
	aiMaster.getChan = make(chan *getAIsReq, 1000)
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
	} else {
		ms.isReady.Unlock()
		reply.Status = mainrpc.OK
		reply.Servers = ms.getServers()
	}

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
	ms.isReady.Lock()
	if !ms.isReady.ready {
		ms.isReady.Unlock()
		reply.Status = paxosrpc.NotReady
		return nil
	}
	ms.isReady.Unlock()

	switch args.Type {
	case paxosrpc.Setup:
		ms.quiese.Lock()
		ms.quiese.ready = false
		ms.quiese.Unlock()

		req := maxRequest{make(chan maxReply)}
		ms.paxosMaster.maxChan <- req
		maxReply := <-req.retChan

		// Wait till done
		<-maxReply.done

		reply.Status = paxosrpc.OK
		reply.Servers = ms.getServers()
		reply.CommandNumber = maxReply.cmdNum

	case paxosrpc.Sync:
		nopReq := paxosrpc.Command{
			CommandNumber: args.CommandNumber,
			Type:          paxosrpc.NOP,
		}

		ms.paxosMaster.commandChan <- nopReq

		// Get Done Chan and wait till done
		req := doneRequest{args.CommandNumber, make(chan chan bool)}
		ms.paxosMaster.doneChan <- req
		done := <-req.retChan

		// Wait till done
		<-done

		reply.Status = paxosrpc.OK

	case paxosrpc.Replace:
		// Get Commands until now
		req := cmdRequest{make(chan []paxosrpc.Command)}
		ms.paxosMaster.getCmdChan <- req
		commands := <-req.retChan

		i := 0
		for j, n := range ms.servers {
			// Update Servers
			if n.hostport == args.ToReplace[i] {
				n.hostport = args.ToAdd[i]
				n.client = dialHTTP(n.hostport)
				ms.servers[j] = n

				// If Master, tell
				if args.Master {
					qArgs := paxosrpc.QuieseArgs{
						Type:     paxosrpc.CatchUp,
						Commands: commands,
						Servers:  ms.getServers(),
					}
					var qReply paxosrpc.QuieseReply
					n.client.Go("PaxosServer.Quiese", &qArgs, &qReply, nil)
				}
			}
		}

		reply.Status = paxosrpc.OK

		ms.quiese.Lock()
		ms.quiese.ready = true
		ms.quiese.Unlock()

	case paxosrpc.CatchUp:
		if !ms.replacement {
			reply.Status = paxosrpc.WrongServer
			return nil
		}
		// Send Commands to commit chan
		for _, c := range args.Commands {
			ms.statsMaster.commitChan <- c
		}

		// Update Server list
		ms.servers = make([]node, len(reply.Servers))

		for i, h := range args.Servers {
			ms.servers[i].hostport = h
		}

		reply.Status = paxosrpc.OK
		// Close ready
		close(ms.ready)
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
