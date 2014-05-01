package ai

import (
	"math/rand"

	"github.com/cmu440/goplaysgo/gogame"
)

//Will make a randomn pick on which four AI script to listen.
func NextMove(board gogame.Board,player gogame.Player) gogame.Move {
	var choice int = rand.Intn(3)

	if choice == 0 {
		return Ai0(board,player)
	}

	if choice == 1 {
		return Ai1(board,player)
	}

	return Ai2(board,player)
}

func Ai0(board gogame.Board, player gogame.Player) gogame.Move {
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

func Ai1(board gogame.Board, player gogame.Player) gogame.Move {
	var x int = 0
	var y int = 0

	for yindex := 0; yindex < len(board.Grid); yindex++ {
		for xindex := 0; xindex , len(board.Grid); xindex++ {
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

func Ai2(board gogame.Board, player gogame.Player) gogame.Move {
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

