package refereeserver

type RefereeServer interface {
	// PlayGame plays the game of go between the two AI servers
	// passed in the arguments. When the game ends, PlayGame will
	// return the winner of the game. Note that this means that when
	// PlayGame is called, it would be best to have it run
	// asynchronously.
	// TODO: Figure out if rpc calls are concurrent, and multiple
	// RPC calls of the smae function are blocked or not. If it is
	// blocked, we should change the api so that the RefereeServer
	// makes an rpc call to the main server to return the results
	PlayGame(*refereerpc.PlayArgs, *refereerpc.PlayReply) error
}
