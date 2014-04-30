package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cmu440/goplaysgo/goclient"
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
)

var port = flag.Int("port", 9099, "Mainserver port number")

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
		fmt.Fprintln(os.Stderr, "  SubmitAI:       sa aiName pathToFile")
		fmt.Fprintln(os.Stderr, "  GetStandings:   sd")
		fmt.Fprintln(os.Stderr, "  GetStats:       st aiName")
		fmt.Fprintln(os.Stderr, "  GetServers:     sv")
	}
}

func main() {
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
		}
	case "sv":
		println("Getting Servers")
		reply, err := client.GetServers()
		printStatus(ci.funcname, reply.Status, err)
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

func printResults(results []mainrpc.GameResult) {
	for _, result := range results {
		fmt.Println("Game:", result.Player1, "vs", result.Player2,
			"[", result.Points1, ":", result.Points2, "]")
	}
}
