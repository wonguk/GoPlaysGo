package ai

import (
	"math/rand"

	"github.com/cmu440/goplaysgo/gogame"
)

// NextMove basically chooses a random empty spot
func NextMove(board gogame.Board, player gogame.Player) gogame.Move {
	var x int = rand.Intn(len(board.Grid))
	var y int = rand.Intn(len(board.Grid))

	for board.Grid[y][x].Player != "" {
		x = rand.Intn(len(board.Grid))
		y = rand.Intn(len(board.Grid))
	}

	return gogame.Move{
		YPos: y,
		XPos: x,
	}
}
