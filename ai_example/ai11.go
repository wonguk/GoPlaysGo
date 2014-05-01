package ai

import (
	"math/rand"

	"github.com/cmu440/goplaysgo/gogame"
)
//Will make a randomn pick on which three beter AI scripts to listen.
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
	var x int = 0
	var y int = 0
	var maxT int = 0

	for yindex := y; yindex < len(board.Grid); yindex++ {
		for xindex := x; xindex < len(board.Grid); xindex++ {
			if (board.IsLegalMove(player,gogame.Move{y,x}) == 1) {
				TestBoard := gogame.MakeBoard(len(board.Grid))
				TestBoard.Passed = board.Passed
				TestBoard.Turn = board.Turn
				TestBoard.Grid = board.Grid
				TestBoard.MakeMove(player,gogame.Move{y,x})
				EmptyChain := make([]int, len(board.Grid)*len(board.Grid))
				StoneChain, ChainLen := TestBoard.FindStoneChain(gogame.Move{yindex, xindex}, TestBoard.Grid[yindex][xindex].Player, EmptyChain, 0)
				TerrorityCount := TestBoard.CountTerritory(StoneChain, ChainLen)
				if TerrorityCount > maxT {
					x = xindex
					y = yindex
					maxT = TerrorityCount
				}
			}
		}
	}
	return gogame.Move{
		YPos: y,
		XPos: x,
	}
}

func Ai1(board gogame.Board, player gogame.Player) gogame.Move {
	var x int = 0
	var y int = 0
	var OpponentX int = 0
	var OpponentY int = 0
	var maxT int = 0
	var Opponent gogame.Player

	if (player == "White") {
		Opponent = "Black"
	} else {
		Opponent = "White"
	}

	for yindex := y; yindex < len(board.Grid); yindex++ {
		for xindex := x; xindex < len(board.Grid); xindex++ {
			if (board.Grid[y][x].Player == Opponent) {
			EmptyChain := make([]int, len(board.Grid)*len(board.Grid))
			StoneChain, ChainLen := board.FindStoneChain(gogame.Move{yindex, xindex}, board.Grid[yindex][xindex].Player, EmptyChain, 0)
			TerrorityCount := board.CountTerritory(StoneChain, ChainLen)
			if TerrorityCount > maxT {
				OpponentX = xindex
				OpponentY = yindex	
				}
			}
		}
	}

	if (board.IsLegalMove(player,gogame.Move{OpponentY-1,OpponentX}))==1 {
		return gogame.Move{
		YPos: OpponentY-1,
		XPos: OpponentX,
		}
	}

	if (board.IsLegalMove(player,gogame.Move{OpponentY+1,OpponentX}))==1 {
		return gogame.Move{
		YPos: OpponentY+1,
		XPos: OpponentX,
		}
	}

	if (board.IsLegalMove(player,gogame.Move{OpponentY,OpponentX+1}))==1 {
		return gogame.Move{
		YPos: OpponentY,
		XPos: OpponentX+1,
		}
	}

	if (board.IsLegalMove(player,gogame.Move{OpponentY,OpponentX-1}))==1 {
		return gogame.Move{
		YPos: OpponentY,
		XPos: OpponentX-1,
		}
	}

	//No viable moves (somehow) so just pass (Make sure it catches this to pass)
	return gogame.Move{
		YPos: 0,
		XPos: 0,
		}
	}

func Ai2(board gogame.Board,player gogame.Player) gogame.Move {
	var LastX int
	var LastY int

	for yindex := 0; yindex < len(board.Grid); yindex++ {
		for xindex := 0; xindex < len(board.Grid); xindex++ {
			if (board.Grid[yindex][xindex].Turn == board.Turn-1) {
					LastX = xindex
					LastY = yindex
				}
			}
		}

	if (board.IsLegalMove(player,gogame.Move{LastY-1,LastX}))==1 {
		return gogame.Move{
		YPos: LastY-1,
		XPos: LastX,
		}
	}

	if (board.IsLegalMove(player,gogame.Move{LastY+1,LastX}))==1 {
		return gogame.Move{
		YPos: LastY+1,
		XPos: LastX,
		}
	}

	if (board.IsLegalMove(player,gogame.Move{LastY,LastX+1}))==1 {
		return gogame.Move{
		YPos: LastY,
		XPos: LastX+1,
		}
	}

	if (board.IsLegalMove(player,gogame.Move{LastY,LastX-1})) ==1{
		return gogame.Move{
		YPos: LastY,
		XPos: LastX-1,
		}
	}

		return gogame.Move{
		YPos: 0,
		XPos: 0,
		}
	}
