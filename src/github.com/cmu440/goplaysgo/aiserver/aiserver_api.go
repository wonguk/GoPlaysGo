package aiserver

type AIServer interface {
	//TODO: Do we need a Referee Server.....?
	// We could have the AIServers implement the GO rules, which would
	// remove the need to have a Referee.
	// This would:
	// 1) Remove the need to have a Referee (Yay less rpc calls)
	// 2) Remove ethe need to send the GO board around (Less Network)

	// Original Implementation:
	// NextMove should return the next move given the current state of
	// the board.
	NextMove(*airpc.NextMoveArgs, *airpc.NextMoveReply) error

	// No Referee Implementation:
	// InitGame Init Game should be called by the MainServer as a
	// 2PC between the two AI servers to check that the two AIs are
	// ready to play each other
	InitGame(*airpc.InitGameArgs, *airpc.InitGameReply) error

	// StartGame should be called on the AI server that will play first
	// in the game. From hereon, the AI servers will communicate with
	// each other, and once they are done, return the results to the
	// main server.
	StartGame(*airpc.StartGameArgs, *airpc.StartGameReply) error

	//TODO decide which AI server should reply to the main server
	// Choices: 1) The server that wins (require separate rpc to main_
	//          2) The Server that started (may return result)
}
