package goclient

import (
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
)

// GoClient is the interface that the GoClient implements
type GoClient interface {
	SubmitAI(string, string) (mainrpc.SubmitAIReply, error)
	GetStandings() (mainrpc.GetStandingsReply, error)
	GetServers() (mainrpc.GetServersReply, error)
}
