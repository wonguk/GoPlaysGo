package airpc

import "github.com/cmu440/goplaysgo/rpc/mainrpc"

type Status int

const (
	OK Status = iota + 1
	NotReady
	InvalidMove
	GameExists
)

// Basic idea behind this is that each AI server
// maintains the board of the game between a given opponent,
// so we only need to send the moves we make back and forth
type NextMoveArgs struct {
	Player string
	XPos   int
	YPos   int
}

type NextMoveReply struct {
	Status Status
	XPos   int
	YPos   int
}

type CheckArgs struct {
	Player string
}

type CheckReply struct {
	Status Status
}

// I think we may need another RPC call before InitGame for the 2PC
type InitGameArgs struct {
	size int
}

type IntGameReply struct {
	Status Status
}

//No args needed
type StartGameArgs struct {
}

//StartGame should reply with the
type StartGameReply struct {
	Status Status
	Result mainrpc.GameResult
}
