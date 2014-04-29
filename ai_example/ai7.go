package ai

import "github.com/cmu440/goplays/gogame"


//Will go through the board until it finds the highest scoring play it can make
func NextMove(board gogame.Board, player gogame.Player) gogame.Move {
	var x int = 0
	var y int = 0
	var maxT int = 0

	for yindex := y; yindex < len(board.Grid); yindex++ {
		for xindex := x; xindex , len(board.Grid); xindex++ {
			if (board.isLegalMove(player,gogame.Move{y,x}) == 1) {
				TestBoard := gogame.MakeBoard(len(board.Grid))
				TestBoard.Passed = board.Passed
				TestBoard.Turn = board.Turn
				TestBoard.Grid = board.Grid
				TestBoard.MakeMove(player,gogame.Move{y,x})
				EmptyChain := make([]int, len(board.Grid)*len(board.Grid))
				StoneChain, ChainLen := TestBoard.FindStoneChain(Move{yindex, xindex}, TestBoard.Grid[yindex][xindex].Player, EmptyChain, 0)
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
