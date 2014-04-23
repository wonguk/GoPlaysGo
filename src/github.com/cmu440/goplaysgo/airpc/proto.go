package airpc

import "github.com/cmu440goplaysgo/gogame"

type GameState struct {
	//This is just like gogame Board, might want to include more
	Turn int
	Grid [][]gogame.Stones
}

type Player struct {

}

type NextMoveArgs {
	Board gogame.Board
	Player string
}

type NextMoveReply{
	//If we are using refree then we need:
	ypos int
	xpos int
	//IF we are not then we just need:
	Board gogame.Board
}

type InitGameArgs {
	//This is needed 
	size int
}

type IntGameReply {
	Board gogame.Board
}

type StartGameArgs {
	size int
}

type StartGameReply {
	Board gogame.Board
	Player string
}