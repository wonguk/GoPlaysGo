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

func main() {
	tests := []testFunc{
		{"testNormalSingle", testNormalSingle},
		{"testNormalMultiple", testNormalMultiple},
		{"testDuplicateSingle", testDuplicateSingle},
		{"testDuplicateMultiple", testDuplicateMultiple},
	}

	flag.Var(&hps, "port", "List of Hosts")
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
	s := make([]goclient.GoClient, len(hps))

	for i, h := range hps {
		c, err := goclient.NewGoClientHP(h)
		if err != nil {
			os.Exit(-1)
		}
		s[i] = c
	}

	return s
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
	time.Sleep(time.Second)
	checkStandings(c)

	c.SubmitAI(getNextAI(), "ai_example/ai_random.go")
	time.Sleep(time.Second)
	checkStandings(c)

	c.SubmitAI(getNextAI(), "ai_example/ai_random.go")
	time.Sleep(time.Second)
	checkStandings(c)

	c.SubmitAI(getNextAI(), "ai_example/ai_random.go")
	time.Sleep(time.Second)
	checkStandings(c)

	time.Sleep(time.Second)
}

func testNormalMultiple() {
	for _, c := range mt.servers {
		c.SubmitAI(getNextAI(), "ai_example/ai_random.go")
		time.Sleep(time.Second)
		for _, s := range mt.servers {
			checkStandings(s)
		}
	}
}

func testDuplicateSingle() {
	c := mt.servers[0]

	reply, err := c.SubmitAI(getCurAI(), "ai_example/ai_random.go")

	if err != nil {
		return
	}

	if reply.Status != mainrpc.AIExists {
		LOGE.Println("There are Duplicate AIs in Server")
		failCount++
		return
	}

	passCount++
	fmt.Println("PASS")
}

func testDuplicateMultiple() {
	for _, c := range mt.servers {
		reply, err := c.SubmitAI(getCurAI(), "ai_example/ai_random.go")

		if err != nil {
			return
		}

		if reply.Status != mainrpc.AIExists {
			LOGE.Println("There are Duplicate AIs in Server")
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
