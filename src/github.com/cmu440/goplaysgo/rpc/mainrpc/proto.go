package mainrpc

import "net/rpc"

// Status represents the Status of a given RPC Call
type Status int

// The different Statuses the MainServer can return
const (
	OK Status = iota + 1
	NotReady
	WrongServer
	AIExists
)

//TODO: Timing constraints for main server?
const ()

// GameResult contains the result of a game between two AIs
type GameResult struct {
	player1 string
	player2 string

	points1 int
	points2 int
}

// Stats stores all the information for a given AI
type Stats struct {
	Name        string
	Hostport    string
	Wins        int
	Losses      int
	Draws       int
	GameResults []GameResult
}

// Standings stores all the Stats for the different AIs
type Standings []Stats

// RegisterArgs contains the host data to register to the master main server
type RegisterArgs struct {
	hostname string
}

// RegisterReply returns all the servers in the paxos chain
type RegisterReply struct {
	Status  Status
	Servers []string
}

// GetServersArgs is empty
type GetServersArgs struct {
}

// GetServersReply returns the servers in the paxos ring
type GetServersReply struct {
	Status  Status
	Servers []string
}

// RegisterRefArgs registers the referee server (DEPRECATED)
type RegisterRefArgs struct {
	hostname string
}

// RegisterRefReply returns the result of adding the ref (DEPRECATED)
type RegisterRefReply struct {
	Status Status
}

// SubmitAIArgs has the AI name and the code that is submitted
type SubmitAIArgs struct {
	name string
	code []byte
}

// SubmitAIReply returns the status of adding the enw AI
type SubmitAIReply struct {
	Status Status
}

// GetStandingsArgs requests the standings
type GetStandingsArgs struct {
}

// GetStandingReply returns the current standings of the AIs
// Note: They are not in any specific order
type GetStandingReply struct {
	Status    Status
	Standings Standings
}
