package paxosrpc

import "github.com/cmu440/goplaysgo/rpc/mainrpc"

type Status int

const (
	OK Status = iota + 1
	NotReady
	Reject
)

const (
	PaxosTimeout = 10
)

type Type int

const (
	Init Type = iota + 1
	Update
	NOP
)

type Command struct {
	CommandNumber int
	Type          Type
	Player        string
	Hostport      string
	GameResult    mainrpc.GameResult
}

type PrepareArgs struct {
	N             int
	CommandNumber int
}

type PrepareReply struct {
	Status    Status
	N         int
	MaxCmdNum int
	Command   Command
}

type AcceptArgs struct {
	N       int
	Command Command
}

type AcceptReply struct {
	Status Status
}

type CommitArgs struct {
	N       int
	Command Command
}

type CommitReply struct {
	Status Status
}
