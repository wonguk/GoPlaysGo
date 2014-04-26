package goclient

import (
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
	"io/ioutil"
	"net"
	"net/rpc"
	"strconv"
)

type goClient struct {
	client *rpc.Client
}

// NewGoClient returns a client for the MainServer in GoPlaysGo
func NewGoClient(hostname string, port int) (*goClient, error) {
	cli, err := rpc.DialHTTP("tcp", net.JoinHostPort(hostname, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}

	return &goClient{client: cli}, nil
}

func (gc *goClient) SubmitAI(name string, path string) (mainrpc.SubmitAIReply, error) {
	var reply mainrpc.SubmitAIReply

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return reply, err
	}

	args := &mainrpc.SubmitAIArgs{
		Name: name,
		Code: b,
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

func (gc *goClient) GetServers() (mainrpc.GetServersReply, error) {
	var args mainrpc.GetServersArgs
	var reply mainrpc.GetServersReply

	err := gc.client.Call("MainServer.GetServers", &args, &reply)

	return reply, err
}
