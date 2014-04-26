package goclient

import (
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
	"io/ioutil"
	"net"
	"rpc"
	"strconv"
)

type goClient struct {
	client *rpc.Client
}

// NewGoClient returns a client for the MainServer in GoPlaysGo
func NewGoClient(hostname string, port int) (*goClient, err) {
	cli, err := rpc.DialHTTP("tcp", net.JoinHostPort(hostname, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}

	return &goClient{client: cli}, nil
}

func (gc *goClient) SubmitAI(name string, path string) (mainrpc.SubmitAIreply, error) {
	var reply mainrpc.SubmitAIReply

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return reply, err
	}

	args := &mainrpc.SubmitAIArgs{
		name: name,
		code: b,
	}

	err = gc.client.Call("MainServer.SubmitAI", args, &reply)

	return reply, err
}

func (gc *goClient) GetStandings() (mainrpc.GetStandingsReply, error) {
	var args mainrpc.GetStandingsArgs
	var reply mainrpc.GetStandingsReply

	err := gc.client.Call("MainServer.GetStandings", &args, &reply)

	return reply, err
}

func (gc *goClient) getServers() (mainrpc.GetServers, error) {
	var args mainrpc.GetServerArgs
	var reply mainrpc.GetServerReply

	err := gc.client.Call("MainServer.GetServers", &args, &reply)

	return reply, err
}
