package mainserver

import "github.com/cmu440/goplaysgo/rpc/paxosrpc"

// PaxosServer will be used in a mainserver so multiple mainservers can
// keep their StatsMasters in sync
type PaxosServer interface {
	// Propose proposes a command to a server in the paxos ring
	Prepare(*paxosrpc.PrepareArgs, *paxosrpc.PrepareReply)

	// Accept sends an accept message to the servers in the paxos ring
	Accept(*paxosrpc.AcceptArgs, *paxosrpc.AcceptReply)

	// Commit Sends a message to a server in the paxos ring to update a given value
	Commit(*paxosrpc.CommitArgs, *paxosrpc.CommitReply)
}
