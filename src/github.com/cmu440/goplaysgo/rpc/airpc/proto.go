package airpc

import (
	"github.com/cmu440/goplaysgo/gogame"
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
)

// Status represents the Status of a given RPC Reply
type Status int

// The Different Possible Statuses
const (
	OK Status = iota + 1
	NotReady
	InvalidMove
	GameExists
)

// NextMoveArgs represents a move that the opponent played
// Basic idea behind this is that each AI server
// maintains the board of the game between a given opponent,
// so we only need to send the moves we make back and forth
type NextMoveArgs struct {
	Player string
	Move   gogame.Move
}

// NextMoveReply represents the move that the AIServer played
type NextMoveReply struct {
	Status Status
	Move   gogame.Move
}

// CheckArgs basically asks the AIServer if it is OK to play the game
// with the player specified
type CheckArgs struct {
	Player string
}

// CheckReply represents whether or not the AIServer is okay to play
type CheckReply struct {
	Status Status
}

// InitGameArgs are the arguments needed to initalize a game with the given opponent
type InitGameArgs struct {
	Player   string
	Hostport string

	Size int
}

// InitGameReply returns the result of initializing a game
type InitGameReply struct {
	Status Status
}

// StartGameArgs starts the game with the specified Player
type StartGameArgs struct {
	Player string
}

//StartGameReply returns the result of the game
type StartGameReply struct {
	Status Status
	Result mainrpc.GameResult
}

type AIPlayer struct {
	Player   string
	Hostport string
}

type StartGamesArgs struct {
	Opponents []AIPlayer
	Servers   []string
}

type StartGamesReply struct {
	Status Status
}
