package ai

import (
	"math/rand"

	"github.com/cmu440/goplays/gogame"
)

//Much like ai1 but will check if the spot is not just empty put also legal (hense no unnessary "passing")
func NextMove(board gogame.Board, player gogame.Player) gogame.Move {
	var x int = rand.Intn(len(board.Grid))
	var y int = rand.Intn(len(board.Grid))
	//Note still want to check if there even is a legal move left
	for (board.isLegalMove(player, gogame.Move{y, x}) != 1) {
		x = rand.Intn(len(board.Grid))
		y = rand.Intn(len(board.Grid))
	}

	return gogame.Move{
		YPos: y,
		XPos: x,
	}
}
