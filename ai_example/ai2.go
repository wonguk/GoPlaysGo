package ai

import (
	"github.com/cmu440/goplaysgo/gogame"
)

//This ai will just go through all the positions on the board until it finds a empty one
func NextMove(board gogame.Board, player gogame.Player) gogame.Move {
	var x int = 0
	var y int = 0

	for yindex := 0; yindex < len(board.Grid); yindex++ {
		for xindex := 0; xindex < len(board.Grid); xindex++ {
			if board.Grid[yindex][xindex].Player != "" {
				x = xindex
				y = yindex
				break
			}
		}
	}

	return gogame.Move{
		YPos: y,
		XPos: x,
	}
}
