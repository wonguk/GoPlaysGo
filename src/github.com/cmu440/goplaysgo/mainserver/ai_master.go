package mainserver

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"net/rpc"
	"os/exec"
	"strconv"
	"time"

	"github.com/cmu440/goplaysgo/gogame"
	"github.com/cmu440/goplaysgo/rpc/airpc"
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
)

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
			LOGV.Println("AIMaster:", "Recieved AI Program", newAI.name)
			_, ok := am.aiClients[newAI.name]

			if ok {
				LOGV.Println("AIMaster:", newAI.name, "Already Exists")
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
				am.aiClients[newAI.name] = newInfo
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
					newInfo.hostport = hostport
					newInfo.client = dialHTTP(hostport)
					am.aiClients[newAI.name] = newInfo
					newAI.retChan <- true
					initChan <- initRequest{newAI.name, hostport}
					go newInfo.testAI(addChan, am.aiClients)
				}
			}
		}
	}
}

func (ai *aiInfo) initAI() (string, error) {
	LOGV.Println("AIMaster:", "Initializing AI", ai.name)
	//TODO use cmd line to compile and run ai
	// go install github.com/cmu440/goplaysgo/runners/airunner
	// $GOPATH/bin/airunner -port
	err := ioutil.WriteFile("src/github.com/cmu440/goplaysgo/ai/ai_impl.go", ai.code, 0666)

	if err != nil {
		LOGE.Println("AIMaster:", "Failed to write AI", ai.name, err)
		return "", err
	}
	build := exec.Command("go", "install", "github.com/cmu440/goplaysgo/runners/airunner")
	var out bytes.Buffer
	build.Stderr = &out
	err = build.Run()
	if err != nil || len(out.String()) != 0 {
		LOGE.Println("AIMaster:", "error compiling", ai.code, err)
		return "", err
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))

	//Keep trying to start the server (This is because we may have overlapping ports)
	for {
		port := r.Intn(10000) + 10000
		LOGV.Println("AIMaster:", "Starting AIServer", ai.name, "at", port)

		run := exec.Command("bin/airunner", "-name", ai.name,
			"-port", strconv.Itoa(port))
		err = run.Start()

		if err == nil {
			LOGV.Println("AIMaster:", ai.name, "Has Started at", port)
			hostport := "localhost:" + strconv.Itoa(port)
			return hostport, nil
		} else {
			LOGE.Println("AIMaster:", "Failed to Start", ai.name, err)
			return "", err
		}
	}
}

func (ai *aiInfo) testAI(resultChan chan mainrpc.GameResult,
	aiClients map[string]*aiInfo) {
	LOGV.Println("AIMaster:", "Starting tests for", ai.name)
	checkArgs := airpc.CheckArgs{}
	oppCheckArgs := airpc.CheckArgs{ai.name}

	var checkReply airpc.CheckReply
	var oppCheckReply airpc.CheckReply

	initArgs := airpc.InitGameArgs{}
	initArgs.Size = gogame.Small
	oppInitArgs := airpc.InitGameArgs{
		Player:   ai.name,
		Hostport: ai.hostport,
		Size:     gogame.Small,
	}
	var initReply airpc.InitGameReply
	var oppInitReply airpc.InitGameReply

	startArgs := airpc.StartGameArgs{}

	for _, opp := range aiClients {
		if opp.name == ai.name {
			continue
		}

		LOGV.Println("AIMaster:", "Initializing Test bewteen", ai.name, "and", opp.name)
		checkArgs.Player = opp.name

		//2PC between the two AIs to test
		LOGV.Println("AIMaster:", "Checking AI Server Statuses for", ai.name, "and", opp.name)
		err1 := ai.client.Call("AIServer.CheckGame", &checkArgs, &checkReply)
		LOGV.Println("AIMaster:", "Checking AI Server Statuses for", ai.name, "and", opp.name)
		err2 := opp.client.Call("AIServer.CheckGame", &oppCheckArgs, &oppCheckReply)
		LOGV.Println("AIMaster:", "Checking AI Server Statuses for", ai.name, "and", opp.name)

		if err1 != nil || err2 != nil {
			LOGE.Println("testAI:", "failed to match up", ai.name,
				opp.name, err1, err2)
			continue
		}

		if checkReply.Status != airpc.OK || oppCheckReply.Status != airpc.OK {
			LOGE.Println("testAI:", "failed to match up", ai.name,
				opp.name, checkReply.Status, oppCheckReply.Status)
			continue
		}

		LOGV.Println("AIMaster:", "Starting Test between", ai.name, "and", opp.name)

		//if OK, make them play game (Async)
		LOGV.Println("AIMaster:", "Initializing AI Game at", ai.name, "and", opp.name)
		initArgs.Player = opp.name
		initArgs.Hostport = opp.hostport
		err1 = ai.client.Call("AIServer.InitGame", &initArgs, &initReply)
		err2 = opp.client.Call("AIServer.InitGame", &oppInitArgs, &oppInitReply)

		LOGV.Println("AIMaster:", "Starting the Game at", ai.name)
		startArgs.Player = opp.name
		var gameResult airpc.StartGameReply
		call := ai.client.Go("AIServer.StartGame", &startArgs, &gameResult, nil)

		//Launch goroutine to notiff statsmaster when game is done
		go handleGameResult(call, resultChan)
	}
}

func handleGameResult(call *rpc.Call, resultChan chan mainrpc.GameResult) {
	//TODO A Heartbeat to check if the AIs are still alive

	LOGV.Println("AIMaster:", "Waiting for match to end")
	// Wait for the Call to be done
	call = <-call.Done

	LOGV.Println("AIMaster:", "Match Done!")

	reply := call.Reply.(*airpc.StartGameReply)

	//TODO Error Handling
	if call.Error != nil {
		LOGE.Println("AIMaster:", "failed to complete game", call.Error)
		return
	}

	if reply.Status != airpc.OK {
		LOGE.Println("AI Master:", "failed to complete game", reply.Status)
		return
	}

	resultChan <- reply.Result
}
