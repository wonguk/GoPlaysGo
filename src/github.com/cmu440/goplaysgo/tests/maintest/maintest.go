package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/cmu440/goplaysgo/goclient"
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
)

//Make Wrapper to mainserver to access paxosMaster
type mainTester struct {
	servers []goclient.GoClient
}

type hostports []string

func (h *hostports) String() string {
	return fmt.Sprintf("%s", *h)
}

func (h *hostports) Set(value string) error {
	*h = append(*h, value)

	return nil
}

type testFunc struct {
	name string
	f    func()
}

var (
	passCount int
	failCount int
	mt        *mainTester
	hps       hostports
	numAIs    int
	code      []byte
)

var LOGE = log.New(os.Stderr, "", log.Lshortfile|log.Lmicroseconds)

var master = flag.String("port", "localhost:9099", "Hostport of master server")
var aiMap = make(map[string]bool)

func main() {
	tests := []testFunc{
		{"testNormalSingle", testNormalSingle},
		{"testNormalMultiple", testNormalMultiple},
		{"testCompileError", testCompileError},
		{"testDuplicateSingle", testDuplicateSingle},
		{"testDuplicateMultiple", testDuplicateMultiple},
	}

	flag.Var(&hps, "ports", "List of Hosts")
	flag.Parse()

	mt = new(mainTester)
	mt.servers = initClients()

	for _, t := range tests {
		fmt.Println("Running", t.name)
		t.f()
	}

	fmt.Printf("Passed (%d/%d) tests\n", passCount, passCount+failCount)
}

func initClients() []goclient.GoClient {
	c, err := goclient.NewGoClientHP(*master)

	if err != nil {
		os.Exit(-1)
	}

	for {
		reply, err := c.GetServers()

		if err == nil && reply.Status == mainrpc.OK {
			s := make([]goclient.GoClient, len(reply.Servers))

			for i, h := range reply.Servers {
				c, err := goclient.NewGoClientHP(h)
				if err != nil {
					os.Exit(-1)
				}
				s[i] = c
			}

			return s
		} else if err != nil {
			LOGE.Println("GoClient Error:", err)
			os.Exit(-1)
		}

		time.Sleep(time.Second * 2)
	}
}
func checkStandings(c goclient.GoClient) {
	r, err := c.GetStandings()

	if err != nil {
		LOGE.Println("Error while getting standings", err)
		return
	}

	if r.Status != mainrpc.OK {
		LOGE.Println("Get Standings should return OK")
		failCount++
		return
	}

	if len(r.Standings) > numAIs {
		LOGE.Println("Number of Stats don't fit")
		failCount++
		return
	}

	numResults := 0

	for _, s := range r.Standings {
		numResults += len(s.GameResults)
	}

	if numResults > numAIs*(numAIs-1) {
		LOGE.Println("Number of GameResults don't fit")
		failCount++
		return
	}

	passCount++
	fmt.Println("PASS")
}

func testNormalSingle() {
	c := mt.servers[0]

	c.SubmitAI(getNextAI(), "ai_example/ai_random.go")
	c.SubmitAI(getNextAI(), "ai_example/ai_random.go")
	c.SubmitAI(getNextAI(), "ai_example/ai_random.go")
	c.SubmitAI(getNextAI(), "ai_example/ai_random.go")

	time.Sleep(5 * time.Second)
	checkStandings(c)
}

func testNormalMultiple() {
	for _, c := range mt.servers {
		c.SubmitAI(getNextAI(), "ai_example/ai_random.go")
	}

	time.Sleep(10 * time.Second)
	for _, s := range mt.servers {
		checkStandings(s)
	}

}

func testCompileError() {
	c := mt.servers[0]

	reply, err := c.SubmitAI(getNextAI(), "ai_example/ai_compile_error.go")

	if err != nil {
		return
	}

	if reply.Status != mainrpc.CompileError {
		failCount++
		return
	}

	passCount++
	fmt.Println("PASS")
}

func testDuplicateSingle() {
	c := mt.servers[0]

	reply, err := c.SubmitAI(getNextAI(), "ai_example/ai_random.go")
	reply, err = c.SubmitAI(getCurAI(), "ai_example/ai_random.go")

	if err != nil {
		return
	}

	if reply.Status != mainrpc.AIExists {
		LOGE.Println("There are Duplicate AIs in Server", reply.Status)
		failCount++
		return
	}

	passCount++
	fmt.Println("PASS")
}

func testDuplicateMultiple() {
	reply, err := mt.servers[0].SubmitAI(getNextAI(), "ai_example/ai_random.go")
	for i, c := range mt.servers {
		reply, err = c.SubmitAI(getCurAI(), "ai_example/ai_random.go")

		if err != nil {
			continue
		}

		if reply.Status != mainrpc.AIExists {
			r, _ := c.GetStandings()
			LOGE.Println(getCurAI())
			LOGE.Println("There are Duplicate AIs in Server", i, r.Standings)
			failCount++
			return
		}

		passCount++
		fmt.Println("PASS")
	}
}

func getCurAI() string {
	return "test" + strconv.Itoa(numAIs)
}

func getNextAI() string {
	numAIs++
	return "test" + strconv.Itoa(numAIs)
}

/*
func submitAI(c goclient.GoClient, name string, file string) (*mainrpc.SubmitAIReply, error) {
	reply, err := c.SubmitAI(name, file)

	if reply.Status == mainrpc.OK {
		aiMap[name] = true
	} else {
		aiMap[name] = false
	}
}*/
