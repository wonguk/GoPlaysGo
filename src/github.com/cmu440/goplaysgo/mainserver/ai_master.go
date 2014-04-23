package mainserver

type aiInfo struct {
	name     string
	code     []byte
	manage   bool
	hostport string
	client   *rpc.Client
}

type newAIReq struct {
	name     string
	code     []byte
	manage   bool
	hostport string
	retChan  chan bool
}

type aiMaster struct {
	aiChan    chan *newAIReq
	aiClients map[string]*aiInfo
}

func (am *aiMaster) startAIMaster(initChan chan initRequest, addChan chan mainrpc.GameResult) {
	for {
		select {
		case newAI := <-am.aiChan:
			_, ok := am.aiClients[newAI.name]

			if ok {
				newAI.retChan <- false
				continue
			}

			//TODO PAXOS TO ADD AI

			newInfo := new(aiInfo)
			newInfo.name = newAI.name
			newInfo.code = newAI.code
			newInfo.manage = newAI.manage
			if !newInfo.manage {
				newInfo.hostport = newAI.hostport
				newInfo.client = dialHTTP(newAI.hostport)
			} else {
				//TODO: if this mainserver manages the AI, we must:
				//      1) start up the AI (if fail, retChan<-fail)
				//      1b) Paxos and let others know
				//      2) match the AI with all the existing AIs
				//      3) update the Stats on the go
				hostport, err := newInfo.initAI()

				if err != nil {
					newAI.retChan <- false
				} else {
					newAI.retChan <- true
					initChan <- initRequest{newAI.name, hostport}
					go newInfo.testAI(addChan)
				}
			}
		}
	}
}

func (ai *aiInfo) initAI() (string, error) {
	//TODO use cmd line to compile and run ai
	// go install github.com/cmu440/goplaysgo/runners/airunner
	// $GOPATH/bin/airunner -port
	err := ioutil.WriteFile("$GOPATH/src/github.com/cmu440/goplaysgo/aiserver/ai_impl.go", ai.cod, 0666)
	build := exec.Command("go", "install", "github.com/cmu440/goplaysgo/runners/airunner")
	var out bytes.Buffer
	build.Stderr = &out
	err = build.Run()
	if err != nil || len(out.String()) != 0 {
		LOGE.Println("error compiling", err)
		return "", err
	}

	r := rand.New(rand.NewSource(time.Unix()))

	//Keep trying to start the server (This is because we may have overlapping ports)
	for {
		port := r.Intn(10000) + 10000

		run := exec.Command("$GOPATH/bin/airunner", "-port", port)
		err = run.Start()

		if err == nil {
			hostport := "localhost:" + strconv.Itoa(port)
			return hostport, nil
		}
	}
}

func (ai *aiInfo) testAI(resultChan mainrpc.GameResult) {
	checkArgs := airpc.CheckArgs{}
	oppCheckArgs := airpc.CheckArgs{ai.name}

	var checkReply1 airpc.CheckReply
	var checkReply2 airpc.CheckReply

	initArgs := airpc.InitArgs{gogame.Small}
	var initReply airpc.InitReply

	startArgs := airpc.StartArgs{}

	for _, opp := range aiInfos {
		args.Player = opponent.name

		//2PC between the two AIs to test
		err1 := ai.client.Call("AIServer.CheckGame", &args, &reply1)
		err2 := opp.client.Call("AIServer.CheckGame", &oppArgs, &reply2)

		if err1 != nil || err2 != nil {
			LOGE.Println("testAI:", "failed to match up", ai.name,
				opp.name, err1, err2)
			continue
		}

		if reply1.Status != airpc.OK || reply2.Status != airpc.OK {
			LOGE.Println("testAI:", "failed to match up", ai.name,
				opp.name, reply1.Status, reply2.Status)
			continue
		}

		//if OK, make them play game (Async)
		err1 = ai.client.Call("AIServer.InitGame", &initArgs, &initReply)
		err2 = opp.client.Call("AIServer.InitGame", &initArgs, &initReply)

		var gameResult airpc.StartReply
		call := ai.Go("AIServer.StartGame", &startArgs, &gameResult, nil)

		//Launch goroutine to notiff statsmaster when game is done
		go handleGameResult(call, resultChan)
	}
}

func handleGameResult(call *rpc.Call, resultChan chan mainrpc.GameResult) {
	//TODO A Heartbeat to check if the AIs are still alive

	// Wait for the Call to be done
	call = <-call.Done

	//TODO Error Handling
	if call.Error != nil {
		LOGE.Println("AIMaster:", "failed to complete game", call.Error)
		return
	}

	if call.Reply.Status != airpc.OK {
		LOGE.Println("AI Master:", "failed to complete game", call.Reply.Status)
		return
	}

	resultChan <- call.Reply.Result
}
