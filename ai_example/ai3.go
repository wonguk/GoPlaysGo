package ai3

import (
	"math/rand"
	"github.com/cmu440/goplays/gogame"
)

//This ai will start at a randomn position and if not free then move until the next position
func NextMove(board gogame.Board, player gogame.Player) gogame.Move {
	var x int = rand.Intn(len(board.Grid))
	var y int = rand.Intn(len(board.Grid))

	if board.Grid[y][x].Player != "" {
		for yindex := y; yindex < len(board.Grid); yindex++ {
			for xindex := x; xindex , len(board.Grid); xindex++ {
				if board.Grid[yindex][xindex].Player != "" {
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