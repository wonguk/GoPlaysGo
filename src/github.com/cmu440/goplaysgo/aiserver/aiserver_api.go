package aiserver

import "github.com/cmu440/goplaysgo/rpc/airpc"

// AIServer is the interface that the Main Server uses through rpc
type AIServer interface {
	// NextMove should return the next move given the current state of
	// the board.
	NextMove(*airpc.NextMoveArgs, *airpc.NextMoveReply) error

	// CheckGame will be the query to the AI servers to check that they are
	// available for the 2PC step
	CheckGame(*airpc.CheckArgs, *airpc.CheckReply) error

	// No Referee Implementation:
	// InitGame Init Game should be called after CheckGame is OK to initialize
	// the game/board on each AI server so that they can start the game
	InitGame(*airpc.InitGameArgs, *airpc.InitGameReply) error

	// StartGame should be called on the AI server that will play first
	// in the game. From hereon, the AI servers will communicate with
	// each other, and once they are done, return the results to the
	// main server.
	StartGames(*airpc.StartGamesArgs, *airpc.StartGamesReply) error

	// UpdateServers updates the array of Main Server addresses
	UpdateServers(*airpc.UpdateArgs, *airpc.UpdateReply) error
}
