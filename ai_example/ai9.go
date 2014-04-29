package ai

import "github.com/cmu440/goplays/gogame"

//Will always place a piece next to the last piece
func NextMove(board gogame.Board,player gogame.Player) gogame.Move {
	var LastX int
	var LastY int

	for yindex := y; yindex < len(board.Grid); yindex++ {
		for xindex := x; xindex , len(board.Grid); xindex++ {
			if (board.Grid[y][x].Turn == board.Turn-1) {
					LastX = xindex
					LastY = yindex
				}
			}
		}

	if (board.isLegalMove(player,gogame.Move{LastY-1,LastX})) {
		return gogame.Move{
		YPos: LastY-1,
		XPos: LastX,
		}
	}

	if (board.isLegalMove(player,gogame.Move{LastY+1,LastX})) {
		return gogame.Move{
		YPos: LastY+1,
		XPos: LastX,
		}
	}

	if (board.isLegalMove(player,gogame.Move{LastY,LastX+1})) {
		return gogame.Move{
		YPos: LastY,
		XPos: LastX+1,
		}
	}

	if (board.isLegalMove(player,gogame.Move{LastY,LastX-1})) {
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
