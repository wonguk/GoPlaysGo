package ai

import "github.com/cmu440/goplays/gogame"

//Will always pick a move that decreases the Opponent's highest scorring terrority
func NextMove(board gogame.Board, player gogame.Player) gogame.Move {
	//var x int = 0
	//var y int = 0
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
		for xindex := x; xindex , len(board.Grid); xindex++ {
			if (board.Grid[y][x].Player == Opponent) {
			EmptyChain := make([]int, len(board.Grid)*len(board.Grid))
			StoneChain, ChainLen := board.FindStoneChain(Move{yindex, xindex}, board.Grid[yindex][xindex].Player, EmptyChain, 0)
			TerrorityCount := board.CountTerritory(StoneChain, ChainLen)
			if TerrorityCount > maxT {
				OpponentX = xindex
				OpponentY = yindex	
				}
			}
		}
	}

	if (board.isLegalMove(player,gogame.Move{OpponentY-1,OpponentX})) {
		return gogame.Move{
		YPos: OpponentY-1,
		XPos: OpponentX,
		}
	}

	if (board.isLegalMove(player,gogame.Move{OpponentY+1,OpponentX})) {
		return gogame.Move{
		YPos: OpponentY+1,
		XPos: OpponentX,
		}
	}

	if (board.isLegalMove(player,gogame.Move{OpponentY,OpponentX+1})) {
		return gogame.Move{
		YPos: OpponentY,
		XPos: OpponentX+1,
		}
	}

	if (board.isLegalMove(player,gogame.Move{OpponentY,OpponentX-1})) {
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

