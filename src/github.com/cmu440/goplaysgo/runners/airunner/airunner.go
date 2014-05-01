package main

import (
	"flag"
	"log"

	"github.com/cmu440/goplaysgo/aiserver"
)

var (
	name         = flag.String("name", "ai", "name of the specific ai running")
	port         = flag.Int("port", 8088, "port number to listen on")
	mainHostPort = flag.String("main", "localhost:9099", "hostport to one of the main servers")
)

func init() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
}

func main() {
	flag.Parse()

	_, err := aiserver.NewAIServer(*name, *port)
	if err != nil {
		log.Fatalln("Failed to create storage server:", err)
	}

	select {}
}
