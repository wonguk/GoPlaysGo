package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cmu440/goplaysgo/goclient"
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
	"github.com/cmu440/goplaysgo/rpc/paxosrpc"
)

type hostports struct {
	ports []string
}

func (h *hostports) String() string {
	return fmt.Sprintf("%s", *h)
}
func (h *hostports) Set(value string) error {
	h.ports = append(h.ports, value)

	return nil
}

var port = flag.Int("port", 9099, "Mainserver port number")
var isMaster = flag.Bool("master", false, "Whether or not the MainServer will be responsible for sending previous commands to newly added server")
var cmdNum = flag.Int("cmdNum", 0, "The Command Number the main server will be synced to")
var toAdd hostports
var toReplace hostports

type cmdInfo struct {
	cmdline  string
	funcname string
	nargs    int
}

func init() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	flag.Usage = func() {

		fmt.Fprintln(os.Stderr, "The crunner program is a testing tool that creates and runs an instance")
		fmt.Fprintln(os.Stderr, "of the GoClient that can be used to test the mainserver.\n")
		fmt.Fprintln(os.Stderr, "Usage:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Possible commands:")
		fmt.Fprintln(os.Stderr, "  SubmitAI:          sa aiName pathToFile")
		fmt.Fprintln(os.Stderr, "  GetStandings:      sd")
		fmt.Fprintln(os.Stderr, "  GetServers:        sv")
		fmt.Fprintln(os.Stderr, "  Quiese(Setup):     qs")
		fmt.Fprintln(os.Stderr, "  Quiese(Sync):      qy -cmdNum")
		fmt.Fprintln(os.Stderr, "  Quiese(Replace):   qr -master -add -replace")
		fmt.Fprintln(os.Stderr, "For Quiese(Replace), use the -master flag to indicate if the server")
		fmt.Fprintln(os.Stderr, "will be a master while initializing the new servers, and use the")
		fmt.Fprintln(os.Stderr, "-replace and -add flags to list the servers")
		fmt.Fprintln(os.Stderr, "to replace and to add to the paxos ring.")
	}
}

func main() {
	flag.Var(&toAdd, "add", "List of hostports to add to paxos ring")
	flag.Var(&toReplace, "replace", "List of hostports to replace in paxos ring")
	flag.Parse()

	println("Starting CRunner")
	cmd := flag.Arg(0)
	client, err := goclient.NewGoClient("localhost", *port)
	if err != nil {
		log.Fatalln("Failed to create GoClient:", err)
	}

	cmdlist := []cmdInfo{
		{"sa", "MainServer.SubmitAI", 2},
		{"sd", "MainServer.GetStandings", 0},
		{"sv", "MainServer.GetServers", 0},
		{"qs", "PaxosServer.Quiese(Setup)", 0},
		{"qy", "PaxosServer.Quiese(Sync)", 0},
		{"qr", "PaxosServer.Quiese(Replace)", 0},
	}

	cmdmap := make(map[string]cmdInfo)
	for _, j := range cmdlist {
		cmdmap[j.cmdline] = j
	}

	ci, found := cmdmap[cmd]
	if !found {
		fmt.Fprintln(os.Stderr, "ASDAS")
		flag.Usage()
		os.Exit(1)
	}
	if flag.NArg() < (ci.nargs + 1) {
		fmt.Fprintln(os.Stderr, "QWEQWE")
		flag.Usage()
		os.Exit(1)
	}

	switch cmd {
	case "sa":
		println("Submitting AI")
		reply, err := client.SubmitAI(flag.Arg(1), flag.Arg(2))
		printStatus(ci.funcname, reply.Status, err)

	case "sd":
		println("Getting Standings")
		reply, err := client.GetStandings()
		printStatus(ci.funcname, reply.Status, err)
		fmt.Println("NumStats:", len(reply.Standings))
		for _, stats := range reply.Standings {
			fmt.Println("Standings for", stats.Name)
			fmt.Println("HostPort:", stats.Hostport)
			fmt.Println("(W/L/D):", stats.Wins, "/", stats.Losses, "/", stats.Draws)
			printResults(stats.GameResults)
			fmt.Println("\n\n")
		}

	case "sv":
		println("Getting Servers")
		reply, err := client.GetServers()
		printStatus(ci.funcname, reply.Status, err)
		for _, server := range reply.Servers {
			fmt.Println(server)
		}

	case "qs":
		println("Starting Quiese Mode")
		reply, err := client.QuieseSetup()
		printPStatus(ci.funcname, reply.Status, err)
		fmt.Println("Last Command Number:", reply.CommandNumber)
		fmt.Println("Servers in Paxos Ring:")
		for _, server := range reply.Servers {
			fmt.Println(server)
		}

	case "qy":
		println("Syncing Quises to Command Number")
		reply, err := client.QuieseSync(*cmdNum)
		printPStatus(ci.funcname, reply.Status, err)

	case "qr":
		println("Replacing Servers")
		if len(toAdd.ports) == 0 || len(toReplace.ports) == 0 {
			println("there must be at least 1 server to add and 1 server to replace")
			return
		}
		if len(toAdd.ports) != len(toReplace.ports) {
			println("The number of servers being added does not equal the number being replaced")
		}

		reply, err := client.QuieseReplace(*isMaster, toAdd.ports, toReplace.ports)
		printPStatus(ci.funcname, reply.Status, err)
	}
}

func statusToString(status mainrpc.Status) (s string) {
	switch status {
	case mainrpc.OK:
		s = "OK"
	case mainrpc.NotReady:
		s = "NotReady"
	case mainrpc.WrongServer:
		s = "Wrong Server"
	case mainrpc.AIExists:
		s = "AI Exists"
	case mainrpc.CompileError:
		s = "Compilation Error"
	}

	return
}

func printStatus(cmdName string, status mainrpc.Status, err error) {
	if err != nil {
		fmt.Println("ERROR:", cmdName, "got error:", err)
	} else {
		fmt.Println(cmdName, "replied with Status", statusToString(status))
	}
}

func statusPToString(status paxosrpc.Status) (s string) {
	switch status {
	case paxosrpc.OK:
		s = "OK"
	case paxosrpc.NotReady:
		s = "NotReady"
	case paxosrpc.WrongServer:
		s = "Wrong Server"
	case paxosrpc.Reject:
		s = "Rejected"
	}

	return
}

func printPStatus(cmdName string, status paxosrpc.Status, err error) {
	if err != nil {
		fmt.Println("ERROR:", cmdName, "got error:", err)
	} else {
		fmt.Println(cmdName, "replied with Status", statusPToString(status))
	}
}

func printResults(results []mainrpc.GameResult) {
	for _, result := range results {
		fmt.Println("Game:", result.Player1, "vs", result.Player2,
			"[", result.Points1, ":", result.Points2, "]")
	}
}
