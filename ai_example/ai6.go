package ai

import (
	"math/rand"
	"github.com/cmu440/goplaysgo/gogame"
)

//This ai will start at a randomn position and if not free then move until the next LEGAL position
func NextMove(board gogame.Board, player gogame.Player) gogame.Move {
	var x int = rand.Intn(len(board.Grid))
	var y int = rand.Intn(len(board.Grid))

	if (board.IsLegalMove(player,gogame.Move {y,x})==1) {
		for yindex := y; yindex < len(board.Grid); yindex++ {
			for xindex := x; xindex < len(board.Grid); xindex++ {
				if (board.IsLegalMove(player,gogame.Move {y,x})==1) {
					x = xindex
					y = yindex
					break
				}
			}
		}
	}

	return gogame.Move{
		YPos: y,
		XPos: x,
	}
}
