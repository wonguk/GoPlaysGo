package ai

import "github.com/cmu440/goplaysgo/gogame"

//Will always place a piece next to the last piece
func NextMove(board gogame.Board,player gogame.Player) gogame.Move {
	var LastX int
	var LastY int
	var x int = 0
	var y int = 0
	for yindex := 0; yindex < len(board.Grid); yindex++ {
		for xindex := 0; xindex < len(board.Grid); xindex++ {
			if (board.Grid[y][x].Turn == board.Turn-1) {
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

	if (board.IsLegalMove(player,gogame.Move{LastY,LastX-1}))==1 {
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
