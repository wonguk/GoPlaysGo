package mainserver

import "github.com/cmu440/goplaysgo/rpc/mainrpc"

//The MainServer interface implements all the external rpc calls by clients
type MainServer interface {
	// RegisterServer will add the servers into the Paxos ring
	// of main servers. Initially, there will be a master main server,
	// which will wait for all the main servers to connect
	RegisterServer(*mainrpc.RegisterArgs, *mainrpc.RegisterReply) error

	// GetServers returns a list of all main servers that are curently
	// connected in the paxos ring
	GetServers(*mainrpc.GetServersArgs, *mainrpc.GetServersReply) error

	// SubmitAI takes in an AI go program and schedules them to
	SubmitAI(*mainrpc.SubmitAIArgs, *mainrpc.SubmitAIReply) error

	// GetStandings returns the current standings of the different AIs
	// in the server.
	GetStandings(*mainrpc.GetStandingsArgs, *mainrpc.GetStandingsReply) error
}
