package goclient

import (
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
	"github.com/cmu440/goplaysgo/rpc/paxosrpc"
)

// GoClient is the interface that the GoClient implements
// A GoClient can be used to send RPC calls to a main server
type GoClient interface {
	SubmitAI(string, string) (mainrpc.SubmitAIReply, error)
	GetStandings() (mainrpc.GetStandingsReply, error)
	GetServers() (mainrpc.GetServersReply, error)
	QuieseSetup() (paxosrpc.QuieseReply, error)
	QuieseSync(int) (paxosrpc.QuieseReply, error)
	QuieseReplace(bool, []string, []string) (paxosrpc.QuieseReply, error)
}
