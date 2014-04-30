package main

import (
	"flag"
	"log"

	"github.com/cmu440/goplaysgo/mainserver"
)

var (
	port           = flag.Int("port", 9099, "port number to listen on")
	masterHostPort = flag.String("master", "", "Master main sterver host port (if non-empty then this will be a slave main server)")
	numNodes       = flag.Int("N", 1, "the number of nodes in the paxos ring")
	replacement    = flag.Bool("r", false, "Whether or not the server is a replacement")
)

func main() {
	flag.Parse()

	_, err := mainserver.NewMainServer(*masterHostPort, *numNodes, *port, *replacement)
	if err != nil {
		log.Fatalln("Failed to create storage server:", err)
	}

	select {}
}
