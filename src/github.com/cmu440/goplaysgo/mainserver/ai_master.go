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

func (am *aiMaster) startAIMaster() {
	for {
		select {
		case newAI := <-am.aiChan:
			_, ok := am.aiClients[newAI.name]

			if ok {
				newAI.retChan <- false
				continue
			}

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
				err := newInfo.initAI()

				if err != nil {
					newAI.retChan <- false
				} else {
					go newInfo.testAI(newAI.retChan)
				}
			}
		}
	}
}

func (ai *aiInfo) initAI() error {
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
		return err
	}

	r := rand.New(rand.NewSource(time.Unix()))

	for {
		port := r.Intn(10000) + 10000

		run := exec.Command("$GOPATH/bin/airunner", "-port", port)
		err = run.Start()

		if err == nil {
			return nil
		}
	}
}

func (ai *aiInfo) testAI() {
	//TODO for each AI, make them play each other
	resultChan := make(chan mainrpc.GameResult)

	//TODO define args type for ai
	args := "TODO"
	oppArgs := "TODO"

	numGames := 0

	for _, opponent := range aiInfos {
		//TODO 2PC between the two AIs to test
		err1 := ai.client.Call("AIServer.InitGame", args, reply1)
		err2 := opponent.client.Call("AIServer.InitGame", oppArgs, reply2)

		if err1 != nil || err2 != nil {
			LOGE.Println("testAI:", "failed to match up", ai.name,
				opponent.name, err1, err2)
			continue
		}

		if reply1.Status != airpc.OK || reply2.Status != airpc.OK {
			LOGE.Println("testAI:", "failed to match up", ai.name,
				opponent.name, reply1.Status, reply2.Status)
			continue
		}

		//TODO if OK, make them play game (Async)
		numGames += 1
		call1 := ai.Go("AIServer.StartGame", args, reply)

		//TODO Launch goroutine to
		go handleGameResult(call1, ret)
	}

	//TODO wait for returns
	for i := 0; i < numGames; i++ {
		reply := <-ret
		//TODO Notify StatsMaster about game result
	}

}
