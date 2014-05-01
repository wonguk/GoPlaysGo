package aiserver

import (
	//	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"sync"

	"github.com/cmu440/goplaysgo/rpc/airpc"
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
)

var errfile, _ = os.OpenFile("logs/AITest.err", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
var logfile, _ = os.OpenFile("logs/AITest.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
var logfile2, _ = os.OpenFile("logs/AI.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

// Error Log
var LOGE = log.New(errfile, "ERROR [AIServer] ",
	log.Lmicroseconds|log.Lshortfile)

// Verbose Log
var LOGV = log.New(logfile, "VERBOSE [AIServer] ", log.Lmicroseconds|log.Lshortfile)

// Test Log
var LOGT = log.New(logfile2, "VERBOSE [AIServer] ", log.Lmicroseconds|log.Lshortfile)

type aiServer struct {
	name     string
	hostport string

	gm *gameMaster
}

// NewAIServer returns an AIServer that plays games with other AI Servers
func NewAIServer(name string, port int) (AIServer, error) {
	LOGV.Println("NewAIServer:", "Initializing AI Server for", name, "at", port)
	as := new(aiServer)

	as.name = name
	as.hostport = "localhost:" + strconv.Itoa(port)

	gm := new(gameMaster)
	gm.name = name
	gm.hostport = as.hostport
	gm.checkChan = make(chan checkReq, 100)
	gm.initChan = make(chan initReq, 100)
	gm.startGameChan = make(chan string, 100)
	gm.startChan = make(chan startReq)
	gm.moveChan = make(chan moveReq, 100)
	gm.serverChan = make(chan []string, 100)
	gm.resultChan = make(chan mainrpc.GameResult, 100)
	gm.games = make(map[string]*gameHandler)

	gm.serverLock = sync.Mutex{}

	gm.oppClients = make(map[string]*rpc.Client)

	as.gm = gm

	go as.gm.startGameMaster()

	rpc.RegisterName("AIServer", as)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+strconv.Itoa(port))
	if e != nil {
		return nil, e
	}

	go http.Serve(l, nil)

	LOGV.Println("NewAIServer:", "Done Initializing server for", name)

	return as, nil
}

func (as *aiServer) NextMove(args *airpc.NextMoveArgs, reply *airpc.NextMoveReply) error {
	replyChan := make(chan *airpc.NextMoveReply)
	req := moveReq{
		args:      args,
		reply:     reply,
		replyChan: replyChan,
	}

	as.gm.moveChan <- req

	reply = <-replyChan

	return nil
}

func (as *aiServer) CheckGame(args *airpc.CheckArgs, reply *airpc.CheckReply) error {
	LOGV.Println("CheckGame:", "Checking Game Between", as.name, "and", args.Player)
	retChan := make(chan bool)

	req := checkReq{
		name:    args.Player,
		retChan: retChan,
	}

	as.gm.checkChan <- req

	if <-retChan {
		reply.Status = airpc.OK
	} else {
		reply.Status = airpc.GameExists
	}

	return nil
}

func (as *aiServer) InitGame(args *airpc.InitGameArgs, reply *airpc.InitGameReply) error {
	LOGV.Println("InitGame:", "Initializing Game Between", as.name, "and", args.Player)
	retChan := make(chan bool)

	req := initReq{
		name:     args.Player,
		hostport: args.Hostport,
		size:     args.Size,
		retChan:  retChan,
	}

	as.gm.initChan <- req

	<-retChan

	reply.Status = airpc.OK
	return nil
}

func (as *aiServer) StartGames(args *airpc.StartGamesArgs, reply *airpc.StartGamesReply) error {
	LOGT.Println("StartGames:", "Opponents:", args.Opponents)
	LOGT.Println("StartGames:", "Servers:", args.Servers)
	LOGV.Println("StartGames:", "Recieved Start Games Request")
	LOGV.Println("StartGames:", len(args.Opponents), "Opponents and", len(args.Servers), "Servers")
	as.gm.startChan <- startReq{args.Opponents, args.Servers}

	reply.Status = airpc.OK

	return nil
}

func (as *aiServer) UpdateServers(args *airpc.UpdateArgs, reply *airpc.UpdateReply) error {
	LOGT.Println("UpdateServers:", args.Servers)
	as.gm.serverChan <- args.Servers

	reply.Status = airpc.OK
	return nil
}
