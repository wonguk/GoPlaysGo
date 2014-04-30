package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"time"

	"github.com/cmu440/goplaysgo/goclient"
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
	"github.com/cmu440/goplaysgo/rpc/paxosrpc"
)

//Make Wrapper to mainserver to access paxosMaster
type paxosTester struct {
	servers []*rpc.Client
	n       int
	cmdNum  int
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
	pt        *paxosTester
	hps       hostports
)

var LOGE = log.New(os.Stderr, "", log.Lshortfile|log.Lmicroseconds)

var master = flag.String("port", "localhost:9099", "Hostport of master")

func main() {
	tests := []testFunc{
		{"testNormal", testNormal},
		{"testFailPrepare", testFailPrepare},
		{"testFailAccept", testFailAccept},
		{"testCatchUp", testCatchup},
	}

	flag.Var(&hps, "ports", "List of Hosts")
	flag.Parse()

	pt = new(paxosTester)
	pt.servers = initClients()

	for _, t := range tests {
		fmt.Println("Running", t.name)
		t.f()
	}

	fmt.Printf("Passed (%d/%d) tests\n", passCount, passCount+failCount)
}

func initClients() []*rpc.Client {
	c, err := goclient.NewGoClientHP(*master)

	if err != nil {
		os.Exit(-1)
	}

	for {
		reply, err := c.GetServers()

		if err == nil && reply.Status == mainrpc.OK {
			s := make([]*rpc.Client, len(reply.Servers))

			for i, h := range reply.Servers {
				client, err := rpc.DialHTTP("tcp", h)

				if err != nil {
					LOGE.Println("Failed To Connect to Servers", err)
					os.Exit(-1)
				}

				s[i] = client
			}

			return s
		}
	}
}

func testNormal() {
	pt.n++
	pt.cmdNum++

	// Prepare Phase
	pArgs := paxosrpc.PrepareArgs{pt.n, pt.cmdNum}
	var pReply paxosrpc.PrepareReply

	for _, c := range pt.servers {
		err := c.Call("PaxosServer.Prepare", &pArgs, &pReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Servers", err)
			return
		}

		if pReply.Status != paxosrpc.OK {
			LOGE.Println("FAIL: Prepare should return OK")
			failCount++
			return
		}
	}

	// Accept Phase
	cmd := paxosrpc.Command{
		CommandNumber: pt.cmdNum,
		Type:          paxosrpc.Init,
		Player:        "testNormal",
		Hostport:      *master, //Dummy port
	}
	aArgs := paxosrpc.AcceptArgs{
		N:       pt.n,
		Command: cmd,
	}
	var aReply paxosrpc.AcceptReply
	for _, c := range pt.servers {
		err := c.Call("PaxosServer.Accept", &aArgs, &aReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Servers", err)
			return
		}

		if aReply.Status != paxosrpc.OK {
			LOGE.Println("FAIL: Accept should return OK")
			failCount++
			return
		}
	}

	// Commit Phase
	cArgs := paxosrpc.AcceptArgs{
		N:       pt.n,
		Command: cmd,
	}
	var cReply paxosrpc.CommitReply
	for _, c := range pt.servers {
		err := c.Call("PaxosServer.Commit", &cArgs, &cReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Servers", err)
			return
		}

		if cReply.Status != paxosrpc.OK {
			LOGE.Println("FAIL: Commit should return OK")
			failCount++
			return
		}
	}

	// Check Value
	pt.n++
	pArgs.N = pt.n
	for _, c := range pt.servers {
		err := c.Call("PaxosServer.Prepare", &pArgs, &pReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Servers", err)
			return
		}

		if pReply.Status != paxosrpc.Reject {
			LOGE.Println("FAIL: Prepare SHould Reject")
			failCount++
			return
		}

		if pReply.Command != cmd {
			LOGE.Println("FAIL: Prepare should return last commit")
			failCount++
			return
		}
	}

	fmt.Println("PASS")
	passCount++
}

func testFailPrepare() {
	pt.n++
	pt.cmdNum++

	// Prepare Phase
	pArgs := paxosrpc.PrepareArgs{pt.n, pt.cmdNum}
	var pReply paxosrpc.PrepareReply

	for _, c := range pt.servers {
		err := c.Call("PaxosServer.Prepare", &pArgs, &pReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Servers", err)
			return
		}

		if pReply.Status != paxosrpc.OK {
			LOGE.Println("FAIL: Prepare should return OK")
			failCount++
			return
		}
	}

	// Retry Prepare Phase
	pArgs.N--
	for _, c := range pt.servers {
		err := c.Call("PaxosServer.Prepare", &pArgs, &pReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Servers", err)
			return
		}

		if pReply.Status == paxosrpc.OK {
			LOGE.Println("FAIL: Prepare should return Reject")
			failCount++
			return
		}
	}

	fmt.Println("PASS")
	passCount++
}

func testFailAccept() {
	pt.n++
	pt.cmdNum++

	// Prepare Phase
	pArgs := paxosrpc.PrepareArgs{pt.n, pt.cmdNum}
	var pReply paxosrpc.PrepareReply

	for _, c := range pt.servers {
		err := c.Call("PaxosServer.Prepare", &pArgs, &pReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Servers", err)
			return
		}

		if pReply.Status != paxosrpc.OK {
			LOGE.Println("FAIL: Prepare should return OK")
			failCount++
			return
		}
	}

	pt.n++
	pt.cmdNum++

	// New Prepare Phase
	pArgs = paxosrpc.PrepareArgs{pt.n, pt.cmdNum}

	for _, c := range pt.servers {
		err := c.Call("PaxosServer.Prepare", &pArgs, &pReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Servers", err)
			return
		}

		if pReply.Status != paxosrpc.OK {
			LOGE.Println("FAIL: Prepare should return OK")
			failCount++
			return
		}
	}

	// Accept Phase
	cmd := paxosrpc.Command{
		CommandNumber: pt.cmdNum,
		Type:          paxosrpc.Init,
		Player:        "testNormal",
		Hostport:      *master, //Dummy port
	}
	aArgs := paxosrpc.AcceptArgs{
		N:       pt.n - 1,
		Command: cmd,
	}
	var aReply paxosrpc.AcceptReply
	for _, c := range pt.servers {
		err := c.Call("PaxosServer.Accept", &aArgs, &aReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Servers", err)
			return
		}

		if aReply.Status == paxosrpc.OK {
			LOGE.Println("FAIL: Accept should return Reject (There was a prepare with greater n)")
			failCount++
			return
		}
	}

	fmt.Println("PASS")
	passCount++
}

func testCatchup() {
	// No need to test "catching up" with single server
	if len(pt.servers) == 1 {
		fmt.Println("PASS")
		passCount++
		return
	}

	numServers := len(pt.servers)

	pt.n++
	pt.cmdNum++

	cmd1 := paxosrpc.Command{
		CommandNumber: pt.cmdNum,
		Type:          paxosrpc.Init,
		Player:        "testCatchup1",
		Hostport:      *master,
	}

	commitCmd(cmd1, pt.n, pt.cmdNum, pt.servers[0:len(pt.servers)/2+1], numServers)
	pt.n += 30
	pt.cmdNum++

	cmd2 := paxosrpc.Command{
		CommandNumber: pt.cmdNum,
		Type:          paxosrpc.Init,
		Player:        "testCatchup2",
		Hostport:      *master,
	}

	commitCmd(cmd2, pt.n, pt.cmdNum, pt.servers[len(pt.servers)/2:len(pt.servers)], numServers)
	pt.n += 30
	pt.cmdNum++

	cmd3 := paxosrpc.Command{
		CommandNumber: pt.cmdNum,
		Type:          paxosrpc.NOP,
	}

	commitCmd(cmd3, pt.n, pt.cmdNum, pt.servers, numServers)

	// Wait for servers to sync up
	time.Sleep(5 * time.Second)

	pt.n += 30
	checkCommited(cmd1, pt.n, pt.servers)
	pt.n += 30
	checkCommited(cmd2, pt.n, pt.servers)
	pt.n += 30
	checkCommited(cmd3, pt.n, pt.servers)

	fmt.Println("PASS")
	passCount++
}

func checkCommited(cmd paxosrpc.Command, n int, servers []*rpc.Client) {
	pArgs := paxosrpc.PrepareArgs{n, cmd.CommandNumber}
	var pReply paxosrpc.PrepareReply

	for i, c := range servers {
		err := c.Call("PaxosServer.Prepare", &pArgs, &pReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Server", i, err)
			continue
		}

		if pReply.Status == paxosrpc.OK {
			LOGE.Println("FAIL: Prepare should return Reject", pReply.Command, cmd, n, i)
			failCount++
			return
		}

		if pReply.Command != cmd {
			LOGE.Println("FAIL: Command is different", pReply.Command, cmd)
			failCount++
			return
		}
	}

	fmt.Println("PASS")
	passCount++
}

func commitCmd(cmd paxosrpc.Command, n int, numCmd int, servers []*rpc.Client, numServers int) {
	// Prepare Phase
	pArgs := paxosrpc.PrepareArgs{n, numCmd}
	var pReply paxosrpc.PrepareReply

	numDead := 0

	for i, c := range servers {
		err := c.Call("PaxosServer.Prepare", &pArgs, &pReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Server", i, err)
			numDead++
			continue
		}

		if pReply.Status != paxosrpc.OK {
			LOGE.Println("FAIL: Prepare should return OK")
			failCount++
			return
		}
	}

	// Accept Phase
	aArgs := paxosrpc.AcceptArgs{
		N:       n,
		Command: cmd,
	}

	numDead = 0

	var aReply paxosrpc.AcceptReply
	for i, c := range servers {
		err := c.Call("PaxosServer.Accept", &aArgs, &aReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Server", i, err)
			numDead++
			continue
		}

		if aReply.Status != paxosrpc.OK {
			LOGE.Println("FAIL: Accept should return OK")
			failCount++
			return
		}
	}

	// Commit Phase
	cArgs := paxosrpc.AcceptArgs{
		N:       n,
		Command: cmd,
	}
	var cReply paxosrpc.CommitReply
	for i, c := range servers {
		err := c.Call("PaxosServer.Commit", &cArgs, &cReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Server", i, err)
			continue
		}

		if cReply.Status != paxosrpc.OK {
			LOGE.Println("FAIL: Commit should return OK")
			failCount++
			return
		}
	}

	time.Sleep(5 * time.Second)

	// Check Value
	pt.n++
	pArgs.N = pt.n
	for i, c := range servers {
		err := c.Call("PaxosServer.Prepare", &pArgs, &pReply)

		if err != nil {
			LOGE.Println("Failed To Connect to Servers", i, err)
			continue
		}

		if pReply.Status != paxosrpc.Reject {
			LOGE.Println("FAIL: Prepare SHould Reject")
			failCount++
			return
		}

		if pReply.Command != cmd {
			LOGE.Println("FAIL: Prepare should return last commit", cmd, pReply.Command)
			failCount++
			return
		}
	}

	fmt.Println("PASS")
	passCount++

	return
}
