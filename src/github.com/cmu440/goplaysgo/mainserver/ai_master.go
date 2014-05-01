package mainserver

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

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
	retChan  chan mainrpc.Status
}

type getAIsReq struct {
	retChan chan []airpc.AIPlayer
}

type aiMaster struct {
	aiChan     chan *newAIReq
	getChan    chan *getAIsReq
	aiClients  map[string]*aiInfo
	serverChan chan []string
}

// AI Master
// The AI Master is responsible for starting the AI servers and also
// matching up the AI servers by sending them the other AI servers
// they should play against.

func (am *aiMaster) startAIMaster(initChan chan initRequest, servers []string) {
	for {
		select {
		case newAI := <-am.aiChan:
			LOGV.Println("AIMaster:", "Recieved AI Program", newAI.name)
			_, ok := am.aiClients[newAI.name]

			if ok {
				LOGV.Println("AIMaster:", newAI.name, "Already Exists")
				newAI.retChan <- mainrpc.AIExists
				continue
			}

			newInfo := new(aiInfo)
			newInfo.name = newAI.name
			newInfo.code = newAI.code
			newInfo.manage = newAI.manage

			if !newInfo.manage {
				newInfo.hostport = newAI.hostport
				newInfo.client = dialHTTP(newAI.hostport)
				am.aiClients[newAI.name] = newInfo
			} else {
				hostport, err := newInfo.initAI()

				if err != nil {
					newAI.retChan <- mainrpc.CompileError
				} else {
					newInfo.hostport = hostport
					newInfo.client = dialHTTP(hostport)

					done := make(chan bool)
					initChan <- initRequest{newAI.name, hostport, done}

					go newInfo.testAI(done, newAI.retChan, am.getChan, servers)
				}
			}
		case req := <-am.getChan:
			ais := make([]airpc.AIPlayer, len(am.aiClients))

			i := 0
			for _, ai := range am.aiClients {
				ais[i].Player = ai.name
				ais[i].Hostport = ai.hostport
				i++
			}

			req.retChan <- ais
		case s := <-am.serverChan:
			servers = s

			for _, ai := range am.aiClients {
				args := airpc.UpdateArgs{servers}
				var reply airpc.UpdateReply
				ai.client.Go("AIServer.UpdateServers", &args, &reply, nil)
			}
		}
	}
}

// initAI initializes the AI by writing it to the correct directory,
// compiling the AI Server, and launching it
func (ai *aiInfo) initAI() (string, error) {
	LOGV.Println("AIMaster:", "Initializing AI", ai.name)
	// Use cmd line to compile and run AIServer
	gopath := os.Getenv("GOPATH")
	aiPath := gopath + filepath.FromSlash("/src/github.com/cmu440/goplaysgo/ai/ai_impl.go")
	err := ioutil.WriteFile(aiPath, ai.code, 0666)

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

		runnerPath := gopath + filepath.FromSlash("/bin/airunner")
		run := exec.Command(runnerPath, "-name", ai.name, "-port", strconv.Itoa(port))
		err = run.Start()

		if err == nil {
			LOGV.Println("AIMaster:", ai.name, "Has Started at", port)
			hostport := "localhost:" + strconv.Itoa(port)
			return hostport, nil
		}
		LOGE.Println("AIMaster:", "Failed to Start", ai.name, err)
		return "", err
	}
}

// testAI waits until it has been confirmed that the AI has been inserted in the
// paxos ring and then sends the AI server an array of opponents the AIServer
// should play against
func (ai *aiInfo) testAI(done chan bool, retChan chan mainrpc.Status,
	getAIChan chan *getAIsReq, servers []string) {
	// Wait until StatsMaster has confirmed that AI initialized
	if !<-done {
		retChan <- mainrpc.AIExists
		return
	}

	// Get Latest AIs from AIMaster
	aiReq := &getAIsReq{make(chan []airpc.AIPlayer)}
	getAIChan <- aiReq
	aiClients := <-aiReq.retChan

	startArgs := airpc.StartGamesArgs{aiClients, servers}
	var startReply airpc.StartGamesReply

	err := ai.client.Call("AIServer.StartGames", &startArgs, &startReply)

	if err != nil || startReply.Status != airpc.OK {
		retChan <- mainrpc.CompileError
	} else {
		retChan <- mainrpc.OK
	}
}
