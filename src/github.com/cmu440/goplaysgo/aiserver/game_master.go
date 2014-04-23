package aiserver

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

type startReq struct {
	name    string
	retChan chan mainrpc.GameResult
}

type moveReq struct {
	args      *airpc.NextMoveArgs
	reply     *airpc.NextMoveReply
	replyChan chan *airpc.NextMoveReply
}

type gameMaster struct {
	name string

	checkChan chan checkReq
	initChan  chan initReq
	startChan chan startReq
	moveChan  chan moveReq

	games map[string]*gameHandler
}

type gameHandler struct {
	name     string
	opponent string

	startChan chan startReq
	moveChan  chan moveReq

	oppClient *rpc.Client
	game      gogame.Board
}

func (gm *gameMaster) startGameMaster() {
	for {
		select {
		case req := <-gm.checkChan:
			_, ok := gm.games[req.name]
			req.retChan <- ok

		case req := <-gm.initChan:
			newGame := gm.initGame(req.name, req.hostport, req.size)
			games[req.name] = newGame

			go newGame.startGameHandler()

			req.retChan <- true

		case req := <-gm.startChan:
			game, ok := gm.games[req.name]

			if !ok {
				LOGE.Println("GameMaster:", "Game between", req.name, "does not exist!")
				continue
			}

			game.startChan <- req

		case req := <-gm.moveChan:
			game, ok := gm.games[req.name]

			if !ok {
				LOGE.Println("GameMaster:", "Game between", req.name, "does not exist!")
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
	gh.startChan = make(chan startReq)
	gh.moveChan = make(chan moveReq)
	gh.oppClient = initClient(hostport)
	gh.game = gogame.makeBoard(size)

	return gh
}

func (gh *gameHandler) startGameHandler() {
	for {
		select {
		case req := <-gh.moveChan:
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

		case req := <-gh.startChan:
			args := airpc.NextMoveArgs{gh.name}
			var reply airpc.NextMoveReply

			for {
				if gh.game.isDone() {
					break
				}

				// Make Move
				nextMove := ai.NextMove(gh.game, gogame.White)

				// Update Current Board
				gh.game.MakeMove(gogame.White, nextMove)

				if gh.game.isDone() {
					break
				}

				// Send RPC to Opponent for Next Move
				args.Move = nextMove
				err := client.Call("AIServer.NextMove", &args, &reply)

				// Update Board from reply
				gh.game.MakeMove(gogame.Black, reply.Move)
			}

			// Return Game Result
			result := mainrpc.GameResult{
				player1: gh.name,
				player2: gh.opponent,
				points1: gh.game.GetPoints(gogame.White),
				points2: gh.game.GetPoints(gogame.Black),
			}

			req.retChan <- result
		}
	}
}