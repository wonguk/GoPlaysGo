package goclient

import (
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
)

type GoClient interface {
	SubmitAI(string, string) (mainrpc.SubmitAIreply, error)
	GetStandings() (mainrpc.GetStandingsReply, error)
	GetServers() (mainrpc.GetServers, error)
}
