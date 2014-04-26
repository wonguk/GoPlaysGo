package airpc

import (
	"github.com/cmu440/goplaysgo/gogame"
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
)

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
	Move   gogame.Move
}

type NextMoveReply struct {
	Status Status
	Move   gogame.Move
}

type CheckArgs struct {
	Player string
}

type CheckReply struct {
	Status Status
}

// I think we may need another RPC call before InitGame for the 2PC
type InitGameArgs struct {
	Player   string
	Hostport string

	Size int
}

type IntGameReply struct {
	Status Status
}

type StartGameArgs struct {
	Player string
}

//StartGame should reply with the
type StartGameReply struct {
	Status Status
	Result mainrpc.GameResult
}
