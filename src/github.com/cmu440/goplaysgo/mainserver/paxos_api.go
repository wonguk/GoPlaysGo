package mainserver

import "github.com/cmu440/goplaysgo/rpc/paxosrpc"

// PaxosServer will be used in a mainserver so multiple mainservers can
// keep their StatsMasters in sync through Paxos
type PaxosServer interface {
	// Propose proposes a command to a server in the paxos ring
	Prepare(*paxosrpc.PrepareArgs, *paxosrpc.PrepareReply)

	// Accept sends an accept message to the servers in the paxos ring
	Accept(*paxosrpc.AcceptArgs, *paxosrpc.AcceptReply)

	// Commit Sends a message to a server in the paxos ring to update a given value
	Commit(*paxosrpc.CommitArgs, *paxosrpc.CommitReply)

	// Quiese supports manually replacing servers
	// The order of operation for Quiese should be:
	// 1) Quiese (Setup): Notify server to reject external requests to modify state
	// 2) Quiese (Sync) : Tells server to sync to a given command number (optional)
	// 3) Quiese (Replace): Tells the servers to replace the given servers and
	//                      add new servers. Also if Quiese Master, it will bring
	//                      the replacement servers into the Paxos ring and send
	//                      all previous commands to the replacement.
	// 4) Quiese (CatchUp): Called by the Quiese Master to the replacement server
	//                      to let the replacement server catch up and join the
	//                      paxos group
	Quiese(*paxosrpc.QuieseArgs, *paxosrpc.QuieseReply)
}
