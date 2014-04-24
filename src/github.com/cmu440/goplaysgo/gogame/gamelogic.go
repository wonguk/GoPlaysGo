package gogame

type Size int

const (
	Small  Size = 9
	Medium Size = 13
	Largs  Size = 19
)

type Player string

const (
	White Player = "White"
	Black Player = "Black"
)

//Board Game structs
type Move struct {
	YPos int
	XPos int
}

type Stones struct {
	Player Player
	Turn   int
}

type Board struct {
	Passed int
	Turn   int
	Grid   [][]Stones
}

//FindPosChain - A helper function that taks a chain of ints and responds with 0
//				 if y and x are in the chain and 1 if they are not
func FindPosChain(chain []int, move Move) int {
	for index := 0; index < len(chain)-1; index += 2 {
		yindex := chain[index]
		xindex := chain[index+1]
		if (move.XPos == xindex) && (move.YPos == yindex) {

			return 0
		}
	}

	return 1
}

//makeBoard - Creates a Board struct with a size by size Grid
func makeBoard(size int) Board {
	Bd := new(Board)
	Bd.Turn = 1
	Bd.Passed = 0
	Bd.Grid = make([][]Stones, size)
	for index := range Bd.Grid {
		Bd.Grid[index] = make([]Stones, size)
	}
	return *Bd
}

//isLegalMove - Checks the board's grid to see if the given x,y pos is safe to place
func (bd *Board) isLegalMove(player Player, move Move) int {
	TestStone := bd.Grid[move.YPos][move.XPos]
	if (TestStone.Player == "White") || (TestStone.Player == "Black") {
		return 0
	}
	return 1
}

//Starting at a certain point on the board this function will branch out and find all connecting pieces
func (bd *Board) FindStoneChain(move Move, player Player, chain []int, count int) ([]int, int) {

	if move.XPos > 0 {
		TempStoneLeft := bd.Grid[move.YPos][move.XPos-1]

		if (TempStoneLeft.Player == player) && (FindPosChain(chain, move) == 1) {
			chain[count] = move.YPos
			count++
			chain[count] = move.XPos - 1
			count++
			chain, count = bd.FindStoneChain(Move{move.YPos, move.XPos - 1}, player, chain, count)
		}
	}
	if len(bd.Grid)-1 > move.XPos {
		TempStoneRight := bd.Grid[move.YPos][move.XPos+1]

		if (TempStoneRight.Player == player) && (FindPosChain(chain, move) == 1) {
			chain[count] = move.YPos
			count++
			chain[count] = move.XPos + 1
			count++
			chain, count = bd.FindStoneChain(Move{move.YPos, move.XPos + 1}, player, chain, count)
		}
	}
	if move.YPos > 0 {
		TempStoneUp := bd.Grid[move.YPos-1][move.XPos]

		if (TempStoneUp.Player == player) && (FindPosChain(chain, move) == 1) {
			chain[count] = move.YPos - 1
			count++
			chain[count] = move.XPos
			count++
			chain, count = bd.FindStoneChain(Move{move.YPos - 1, move.XPos}, player, chain, count)
		}
	}
	if len(bd.Grid)-1 > move.YPos {
		TempStoneDown := bd.Grid[move.YPos+1][move.XPos]

		if (TempStoneDown.Player == player) && (FindPosChain(chain, move) == 1) {
			chain[count] = move.YPos + 1
			count++
			chain[count] = move.XPos
			count++
			chain, count = bd.FindStoneChain(Move{move.YPos + 1, move.XPos}, player, chain, count)
		}
	}
	if FindPosChain(chain, move) == 1 {
		chain[count] = move.YPos
		count++
		chain[count] = move.XPos
		count++
		return chain, count
	}
	return chain, count
}

func (bd *Board) CountTerritory(pieces []int, count int) int {
	TCount := 0
	for index := 0; index < count; index = index + 2 {
		ypos := pieces[index]
		xpos := pieces[index+1]
		if xpos > 0 {
			TempStone := bd.Grid[ypos][xpos-1]
			if TempStone.Player == "" {
				TempStone.Player = "Taken"
				TCount++
				bd.Grid[ypos][xpos-1] = TempStone
			}
		}
		if xpos < len(bd.Grid)-1 {
			TempStone := bd.Grid[ypos][xpos+1]
			if TempStone.Player == "" {
				TempStone.Player = "Taken"
				TCount++
				bd.Grid[ypos][xpos+1] = TempStone
			}
		}
		if ypos > 0 {
			TempStone := bd.Grid[ypos-1][xpos]
			if TempStone.Player == "" {
				TempStone.Player = "Taken"
				TCount++
				bd.Grid[ypos-1][xpos] = TempStone
			}
		}
		if ypos < len(bd.Grid)-1 {
			TempStone := bd.Grid[ypos+1][xpos]
			if TempStone.Player == "" {
				TempStone.Player = "Taken"
				TCount++
				bd.Grid[ypos+1][xpos] = TempStone
			}
		}
	}
	for yindex := 0; yindex < len(bd.Grid); yindex = yindex + 1 {
		for xindex := 0; xindex < len(bd.Grid); xindex = xindex + 1 {
			TempStone := bd.Grid[yindex][xindex]
			if TempStone.Player == "Taken" {
				TempStone.Player = ""
				bd.Grid[yindex][xindex] = TempStone
			}
		}
	}
	return TCount
}

func (bd *Board) makeMove(player Player, move Move, turn int) {
	PlacedStone := new(Stones)
	PlacedStone.Player = player
	PlacedStone.Turn = turn
	if bd.isLegalMove(player, move) == 1 {
		bd.Grid[move.YPos][move.XPos] = *PlacedStone
		for yindex := 0; yindex < len(bd.Grid); yindex++ {
			for xindex := 0; xindex < len(bd.Grid); xindex++ {
				if bd.Grid[yindex][xindex].Player != "" {
					EmptyChain := make([]int, len(bd.Grid)*len(bd.Grid))
					StoneChain, ChainLen := bd.FindStoneChain(Move{yindex, xindex}, bd.Grid[yindex][xindex].Player, EmptyChain, 0)
					TerrorityCount := bd.CountTerritory(StoneChain, ChainLen)
					if TerrorityCount == 0 {
						for index := 0; index < ChainLen*2; index = index + 2 {
							yremove := StoneChain[index]
							xremove := StoneChain[index+1]
							bd.Grid[yremove][xremove].Player = ""
							bd.Grid[yremove][xremove].Turn = 0
						}
					}
				}
			}
		}
		bd.Passed = 0
		bd.Turn++
		return
	}
	bd.Passed++
	bd.Turn++
	return
}

func (bd *Board) isDone() bool {
	return bd.Passed == 2 //Two turns passed with both players passing
}