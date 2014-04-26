package mainserver

import (
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
)

// Error Log
var LOGE = log.New(ioutil.Discard, "ERROR [MainServer] ",
	log.Lmicroseconds|log.Lshortfile)

// Verbose Log
var LOGV = log.New(ioutil.Discard, "VERBOSE [MainServer] ",
	log.Lmicroseconds|log.Lshortfile)

type isReady struct {
	sync.Mutex
	ready bool
}

type node struct {
	hostport string
	client   *rpc.Client
}

type mainServer struct {
	master     bool
	masterLock sync.Mutex
	readyLock  sync.Mutex
	masterAddr string
	numNodes   int
	port       int
	servers    []node

	ready   chan struct{}
	isReady isReady

	aiMaster    *aiMaster
	statsMaster *statsMaster
	//TODO Paxos variables
}

// NewMainServer returns a mainserver that manages the different AIs and
// their stats
func NewMainServer(masterServerHostPort string, numNodes, port int) (MainServer, error) {
	ms := new(mainServer)

	if masterServerHostPort == "" {
		LOGV.Println("Starting Master...:", port)
		ms.master = true
	} else {
		LOGV.Println("Starting Slave...:", port)
		ms.master = false
	}

	//TODO Possibly change later to test distributed impl
	hostport := "localhost:" + strconv.Itoa(port)

	ms.masterLock = sync.Mutex{}
	ms.readyLock = sync.Mutex{}
	ms.masterAddr = masterServerHostPort
	ms.numNodes = numNodes
	ms.port = port

	ms.servers = []node{}

	ms.ready = make(chan struct{})

	statsMaster := new(statsMaster)
	statsMaster.reqChan = make(chan statsRequest)
	statsMaster.allReqChan = make(chan allStatsRequest)
	statsMaster.initChan = make(chan initRequest)
	statsMaster.addChan = make(chan mainrpc.GameResult)
	statsMaster.stats = make(map[string]mainrpc.Stats)

	go statsMaster.startStatsMaster()

	ms.statsMaster = statsMaster

	aiMaster := new(aiMaster)
	aiMaster.aiChan = make(chan *newAIReq)
	aiMaster.aiClients = make(map[string]*aiInfo)

	go aiMaster.startAIMaster(ms.statsMaster.initChan, ms.statsMaster.addChan)

	ms.aiMaster = aiMaster

	rpc.RegisterName("MainServer", ms)
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
			ms.isReady.Lock()
			ms.isReady.ready = true
			ms.isReady.Unlock()
			return ms, nil
		}

		LOGV.Println("Master:", port, "waiting for nodes to register")
		<-ms.ready

		ms.isReady.Lock()
		ms.isReady.ready = true
		ms.isReady.Unlock()

		return ms, nil
	}

	LOGV.Println("Slave:", port, "Dialing master")
	client := dialHTTP(masterServerHostPort)

	args := mainrpc.RegisterArgs{hostport}
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

			ms.isReady.Lock()
			ms.isReady.ready = true
			ms.isReady.Unlock()

			return ms, nil
		}

		if err != nil {
			LOGE.Println("Slave:", port, "erro registereing to master", err)
		}

		LOGV.Println("Slave:", port, "sleeping for 1 second")
		time.Sleep(time.Second)
	}

	return nil, errors.New("should have been successful")
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

func (ms *mainServer) getServers() []string {
	servers := make([]string, len(ms.servers))

	for i, n := range ms.servers {
		servers[i] = n.hostport
	}

	return servers
}

// RegisterReferee will add a given referee to the pool of referees
// TODO Need to decide how to spawn/handle referees
func (ms *mainServer) RegisterReferee(*mainrpc.RegisterRefArgs, *mainrpc.RegisterRefReply) error {

	return errors.New("not implemented")
}

// GetServers returns a list of all main servers that are curently
// connected in the paxos ring
func (ms *mainServer) GetServers(args *mainrpc.GetServersArgs, reply *mainrpc.GetServersReply) error {
	ms.isReady.Lock()
	defer ms.isReady.Unlock()
	if !ms.isReady.ready {
		reply.Status = mainrpc.NotReady
	} else {
		reply.Status = mainrpc.OK
		reply.Servers = ms.getServers()
	}

	return nil
}

// SubmitAI takes in an AI go program and schedules them to
func (ms *mainServer) SubmitAI(args *mainrpc.SubmitAIArgs, reply *mainrpc.SubmitAIReply) error {
	retChan := make(chan bool)
	req := &newAIReq{
		name:     args.Name,
		code:     args.Code,
		manage:   true,
		hostport: "",
		retChan:  retChan,
	}

	ms.aiMaster.aiChan <- req

	if <-retChan {
		reply.Status = mainrpc.OK
	} else {
		reply.Status = mainrpc.AIExists
	}

	return nil
}

// GetStandings returns the current standings of the different AIs
// in the server.
func (ms *mainServer) GetStandings(args *mainrpc.GetStandingsArgs, reply *mainrpc.GetStandingsReply) error {
	retChan := make(chan mainrpc.Standings)

	reply.Standings = <-retChan
	reply.Status = mainrpc.OK

	return nil
}

// TODO: Decide whether or not the RefereeServer should make a rpc
// call to the MainServer to return results of a game.

func (ms *mainServer) initClients() {
	for _, n := range ms.servers {
		n.client = dialHTTP(n.hostport)
	}
}

func dialHTTP(hostport string) *rpc.Client {
	client, err := rpc.DialHTTP("tpc", hostport)

	for err != nil {
		time.Sleep(time.Second)
		client, err = rpc.DialHTTP("tcp", hostport)
	}

	return client
}
