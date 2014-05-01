package ai

import (
	"math/rand"

	"github.com/cmu440/goplaysgo/gogame"
)

// NextMove basically chooses a random empty spot
func NextMove(board gogame.Board, player gogame.Player) gogame.Move {
	var x int = rand.Intn(len(board.Grid))
	var y int = rand.Intn(len(board.Grid))

	limit := 0

	for board.Grid[y][x].Player != "" && limit < 20 {
		x = rand.Intn(len(board.Grid))
		y = rand.Intn(len(board.Grid))
		limit++
	}

	return gogame.Move{
		YPos: y,
		XPos: x,
	}
}
