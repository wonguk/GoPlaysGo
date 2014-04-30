package aiserver

import (
	"net/rpc"
	"time"

	"github.com/cmu440/goplaysgo/ai"
	"github.com/cmu440/goplaysgo/gogame"
	"github.com/cmu440/goplaysgo/rpc/airpc"
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
)

type checkReq struct {
	name    string
	retChan chan bool
}

type initReq struct {
	name     string
	hostport string
	size     int
	retChan  chan bool
}

type moveReq struct {
	args      *airpc.NextMoveArgs
	reply     *airpc.NextMoveReply
	replyChan chan *airpc.NextMoveReply
}

type startReq struct {
	opponents   []airpc.AIPlayer
	mainServers []string
}

type gameMaster struct {
	name     string
	hostport string

	checkChan     chan checkReq
	initChan      chan initReq
	startGameChan chan string
	startChan     chan startReq
	moveChan      chan moveReq

	games map[string]*gameHandler

	mainServers       []string
	mainServerClients []*rpc.Client
	oppClients        map[string]*rpc.Client
}

type gameHandler struct {
	name     string
	opponent string

	startChan chan string
	moveChan  chan moveReq

	oppClient *rpc.Client
	game      gogame.Board
}

type gameStarter struct {
	name        string
	opponents   []airpc.AIPlayer
	oppClients  []*rpc.Client
	mainServers []string
}

func (gm *gameMaster) startGameMaster() {
	for {
		select {
		case req := <-gm.checkChan:
			LOGV.Println("Game Master:", gm.name, "Checking Game Against", req.name)
			_, ok := gm.games[req.name]
			req.retChan <- !ok

		case req := <-gm.initChan:
			LOGV.Println("Game Master:", gm.name, "Initializing Game Against", req.name)
			newGame := gm.initGame(req.name, req.hostport, req.size)
			gm.games[req.name] = newGame

			go newGame.startGameHandler(gm.mainServerClients)

			req.retChan <- true

		case req := <-gm.startGameChan:
			LOGV.Println("Game MAster:", gm.name, "Starting Game Against", req)
			game, ok := gm.games[req]

			if !ok {
				LOGE.Println("GameMaster:", "Game between", req, "does not exist!")
				continue
			}

			game.startChan <- req

		case req := <-gm.startChan:
			LOGV.Println("Game Master:", gm.name, "Starting Games!!")
			gm.mainServers = req.mainServers
			gm.mainServerClients = make([]*rpc.Client, len(gm.mainServers))

			for i, s := range gm.mainServers {
				gm.mainServerClients[i] = initClient(s)
			}

			gs := new(gameStarter)
			gs.name = gm.name
			gs.opponents = req.opponents
			gs.oppClients = make([]*rpc.Client, len(req.opponents))

			for i, opp := range req.opponents {
				gs.oppClients[i] = initClient(opp.Hostport)
				gm.oppClients[opp.Player] = gs.oppClients[i]
			}

			go gs.startGameStarter(gm.hostport, gm.checkChan, gm.initChan, gm.startGameChan)

		case req := <-gm.moveChan:
			game, ok := gm.games[req.args.Player]

			if !ok {
				LOGE.Println("GameMaster:", "Game between", req.args.Player,
					"does not exist!")
				continue
			}

			game.moveChan <- req
		}
	}
}

func (gm *gameMaster) initGame(name string, hostport string, size int) *gameHandler {
	gh := new(gameHandler)
	gh.name = gm.name
	gh.opponent = name
	gh.startChan = make(chan string)
	gh.moveChan = make(chan moveReq)
	client, ok := gm.oppClients[name]

	if !ok {
		client = initClient(hostport)
		gm.oppClients[name] = client
	}

	gh.oppClient = client

	gh.game = gogame.MakeBoard(size)

	return gh
}

func (gs *gameStarter) startGameStarter(hostport string, checkChan chan checkReq, initChan chan initReq, startChan chan string) {
	LOGV.Println("GameStarter:", gs.name, "STARIGN!!")

	for i, opp := range gs.opponents {
		if opp.Player == gs.name {
			continue
		}

		go gs.startGame(i, hostport, checkChan, initChan, startChan, opp)
	}
}
func (gs *gameStarter) startGame(i int, hostport string, checkChan chan checkReq,
	initChan chan initReq, startChan chan string, opp airpc.AIPlayer) {

	cReq := checkReq{retChan: make(chan bool)}
	var checkArgs airpc.CheckArgs = airpc.CheckArgs{gs.name}
	var checkReply airpc.CheckReply

	iReq := initReq{
		size:    gogame.Small,
		retChan: make(chan bool),
	}
	var initArgs airpc.InitGameArgs = airpc.InitGameArgs{gs.name, hostport, gogame.Small}
	var initReply airpc.InitGameReply

	LOGV.Println("GameStarter:", gs.name, "Next Opp:", opp.Player)
	client := gs.oppClients[i]

	if client == nil {
		return
	}

	// Check Phase
	LOGV.Println("GameStarter:", gs.name, "v", opp.Player, "Checking")
	cReq.name = opp.Player
	checkChan <- cReq
	if !<-(cReq.retChan) {
		return
	}

	err := client.Call("AIServer.CheckGame", &checkArgs, &checkReply)

	if err != nil || checkReply.Status != airpc.OK {
		LOGE.Println("GameStarter:", "Failed while checking with",
			opp.Player, "error:", err, "status", checkReply.Status)
		return
	}

	// Init Phase
	LOGV.Println("GameStarter:", gs.name, "v", opp.Player, "Initializing")
	iReq.name = opp.Player
	iReq.hostport = opp.Hostport
	initChan <- iReq
	if !<-(iReq.retChan) {
		return
	}

	err = client.Call("AIServer.InitGame", &initArgs, &initReply)

	if err != nil || initReply.Status != airpc.OK {
		LOGE.Println("GameStarter:", "Failed to Init with",
			opp.Player, "error:", err, "status", initReply.Status)
		return
	}

	//Start Phase
	LOGV.Println("GameStarter:", gs.name, "v", opp.Player, "Starting")
	startChan <- opp.Player
}

func (gh *gameHandler) startGameHandler(mainServers []*rpc.Client) {
	for {
		select {
		case req := <-gh.moveChan:
			LOGV.Println("GameHandler:", "Move Request Recieved!")
			//TODO Error Handling!
			// Update current board to use the move specified
			gh.game.MakeMove(gogame.White, req.args.Move)

			// Make Move
			nextMove := ai.NextMove(gh.game, gogame.Black)

			// Update Current Board
			gh.game.MakeMove(gogame.Black, nextMove)

			// return move made
			req.reply.Status = airpc.OK
			req.reply.Move = nextMove

			req.replyChan <- req.reply

		case <-gh.startChan:
			LOGV.Println("GameHandler:", "Starting a Game Between", gh.name, "and", gh.opponent)

			if gh.oppClient == nil {
				return
			}

			args := airpc.NextMoveArgs{}
			args.Player = gh.name
			var reply airpc.NextMoveReply

			for {
				if gh.game.IsDone() {
					LOGV.Println("GameHandler:", "Game Done!!", gh.name, ",", gh.opponent)
					break
				}

				// Make Move
				nextMove := ai.NextMove(gh.game, gogame.White)
				LOGV.Println("GameHandler:", gh.name, "made move", nextMove.XPos, nextMove.YPos)

				// Update Current Board
				gh.game.MakeMove(gogame.White, nextMove)

				if gh.game.IsDone() {
					break
				}

				// Send RPC to Opponent for Next Move
				args.Move = nextMove
				err := gh.oppClient.Call("AIServer.NextMove", &args, &reply)

				if err != nil {
					LOGE.Println("GameHandler:", "RPC Error against", gh.opponent, err)
				}
				LOGV.Println("GameHandler:", gh.opponent, "made move", reply.Move.XPos, reply.Move.YPos)

				// Update Board from reply
				gh.game.MakeMove(gogame.Black, reply.Move)
			}

			LOGV.Println("GameHandler:", "Returning Result of Game between", gh.name, ",", gh.opponent)
			// Return Game Result
			result := mainrpc.GameResult{
				Player1: gh.name,
				Player2: gh.opponent,
				Points1: gh.game.PlayerPoints(gogame.White),
				Points2: gh.game.PlayerPoints(gogame.Black),
			}

			submitArgs := mainrpc.SubmitResultArgs{result}
			var submitReply mainrpc.SubmitResultReply

			//req.retChan <- result
			for {
				for _, c := range mainServers {
					if c == nil {
						continue
					}
					err := c.Call("MainServer.SubmitResult", &submitArgs, &submitReply)

					if err == nil && submitReply.Status == mainrpc.OK {
						return
					}

					LOGE.Println("Error Calling SubmitResult", err)
					time.Sleep(time.Second)
				}
			}

			return
		}
	}
}

func initClient(hostport string) *rpc.Client {
	client, err := rpc.DialHTTP("tcp", hostport)
	limit := 0

	for err != nil && limit < 20 {
		LOGV.Println("InitClient, retrying", err)
		limit++
		time.Sleep(time.Second)
		client, err = rpc.DialHTTP("tcp", hostport)
	}

	return client
}
