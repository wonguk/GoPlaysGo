package aiserver

import (
	"github.com/cmu440/goplaysgo/gogame"
	"github.com/cmu440/goplaysgo/rpc/airpc"
)

type aiServer struct {
	name     string
	hostport string

	gm *gameMaster
}

func NewAIServer(name string, port int) (AIServer, error) {
	as := new(aiServer)

	as.name = name
	as.hostport = "localhost:" + strconv.Itoa(port) //TODO Change later

	gm := new(gameMaster)
	gm.name = name
	gm.checkChan = make(chan checkReq)
	gm.initChan = make(chan initReq)
	gm.startChan = make(chan startReq)
	gm.moveChan = make(chan moveReq)
	gm.games = make(map[string]*gameHandler)

	as.gm = gm

	go as.gm.startGameMaster()

	rpc.RegisterName("AIServer", as)
	rpc.HandleHTTP()
	l, e := net.Lsiten("tcp", ":"+strconv.Itoa(port))
	if e != nil {
		return nil, e
	}

	go http.Serve(l, nil)

	return as, nil
}

func (as *aiServer) NextMove(args *airpc.NextMoveArgs, reply *airpc.NextMoveReply) error {
	//NOTE: When NextMove is called, always make a BLACK Move
	replyChan := make(chan *airpc.NextMoveReply)
	req := moveReq{
		args:      args,
		reply:     reply,
		replyChan: replyChan,
	}

	as.gameMaster.moveChan <- req

	reply = <-replyChan

	return nil
}

func (as *aiServer) CheckGame(args *airpc.CheckArgs, reply *airpc.CheckReply) error {
	retChan := make(chan bool)

	req := checkReq{
		name:    args.Player,
		retChan: retChan,
	}

	if <-retChan {
		reply.Status = airpc.OK
	} else {
		reply.Status = airpc.GameExists
	}

	return nil
}

func (as *aiServer) InitGame(args *airpc.InitGameArgs, reply *airpc.InitGameReply) error {
	retChan := make(chan bool)

	req := initReq{
		name:     args.Player,
		hostport: args.Hostport,
		size:     args.Size,
		retChan:  retChan,
	}

	<-retChan

	reply.Status = airpc.OK
	return nil
}

func (as *aiServer) StartGame(args *airpc.StartGameArgs, reply *airpc.StartGameReply) error {
	//NOTE: The Player who starts the should make a WHITE Move
	retChan := make(chan mainrpc.GameResult)

	req := startReq{
		name:    args.Player,
		retChan: retChan,
	}

	reply.Status = mainrpc.OK
	reply.Result = <-retChan

	return nil
}
