package gogame

import "fmt"

type Size int

const (
	Small  Size = 9
	Medium Size = 13
	Largs  Size = 19
)

type Stones struct {
	Player string //Should be either Black or White
	Turn   int    //The turn in which the stone was placed.
}

type Board struct {
	Turn int
	Grid [][]Stones
}

func makeBoard(size int) Board {
	Bd := new(Board)
	Bd.Turn = 1
	Bd.Grid = make([][]Stones, size)
	for index := range Bd.Grid {
		Bd.Grid[index] = make([]Stones, size)
	}
	return *Bd
}

func isLegalMove(grid [][]Stones, player string, xpos int, ypos int) int {
	//Check if there is a piece there
	TestStone := grid[ypos][xpos]
	if (TestStone.Player == "White") || (TestStone.Player == "Black") {
		return 0
	}
	return 1
}

func FindPosChain(chain []int, ypos int, xpos int) int {
	//fmt.Println("Called FindPosChain:", chain, xpos, ypos)
	for index := 0; index < len(chain); index += 2 {
		yindex := chain[index]
		xindex := chain[index+1]
		if (xpos == xindex) && (ypos == yindex) {
			//fmt.Println("Found it")
			return 0
		}
	}
	//fmt.Println("Not here")
	return 1
}

func FindStoneChain(grid [][]Stones, ypos int, xpos int, player string, chain []int, count int) ([]int, int) {
	//fmt.Println("Called FindStoneChain at:", xpos, " ", ypos)
	if xpos > 0 {
		TempStoneLeft := grid[ypos][xpos-1]
		//fmt.Println("TempStoneLeft:", TempStoneLeft)
		if (TempStoneLeft.Player == player) && (FindPosChain(chain, ypos, xpos) == 1) {
			//fmt.Println("Turned left")
			chain[count] = ypos
			count++
			chain[count] = xpos
			count++
			chain, count = FindStoneChain(grid, ypos, xpos-1, player, chain, count)
		}
	}
	if len(grid)-1 > xpos {
		TempStoneRight := grid[ypos][xpos+1]
		//fmt.Println("TempStoneRight:", TempStoneRight)
		if (TempStoneRight.Player == player) && (FindPosChain(chain, ypos, xpos) == 1) {
			//fmt.Println("Turned right")
			chain[count] = ypos
			count++
			chain[count] = xpos + 1
			count++
			chain, count = FindStoneChain(grid, ypos, xpos+1, player, chain, count)
		}
	}
	if ypos > 0 {
		TempStoneUp := grid[ypos-1][xpos]
		//fmt.Println("TempStoneUp:", TempStoneUp)
		if (TempStoneUp.Player == player) && (FindPosChain(chain, ypos, xpos) == 1) {
			//fmt.Println("Turned Up")
			chain[count] = ypos - 1
			count++
			chain[count] = xpos
			count++
			chain, count = FindStoneChain(grid, ypos-1, xpos, player, chain, count)
		}
	}
	if len(grid)-1 > ypos {
		TempStoneDown := grid[ypos+1][xpos]
		//fmt.Println("TempStoneDown:", TempStoneDown)
		if (TempStoneDown.Player == player) && (FindPosChain(chain, ypos, xpos) == 1) {
			//fmt.Println("Turned Down")
			chain[count] = ypos + 1
			count++
			chain[count] = xpos
			count++
			chain, count = FindStoneChain(grid, ypos+1, xpos, player, chain, count)
		}
	}
	if FindPosChain(chain, ypos, xpos) == 1 {
		chain[count] = ypos
		count++
		chain[count] = xpos
		count++
		return chain, count
	}
	return chain, count
}

func CountTerritory(grid [][]Stones, pieces []int, count int) int {
	TCount := 0
	//fmt.Println("Called Count Territory with:", pieces)
	for index := 0; index < count; index = index + 2 {
		ypos := pieces[index]
		xpos := pieces[index+1]
		if xpos > 0 {
			//fmt.Println("Here with:", ypos, xpos)
			TempStone := grid[ypos][xpos-1]
			if TempStone.Player == "" {
				TempStone.Player = "Taken"
				TCount++
				grid[ypos][xpos-1] = TempStone
			}
		}
		if xpos < len(grid)-1 {
			TempStone := grid[ypos][xpos+1]
			if TempStone.Player == "" {
				TempStone.Player = "Taken"
				TCount++
				grid[ypos][xpos+1] = TempStone
			}
		}
		if ypos > 0 {
			TempStone := grid[ypos-1][xpos]
			if TempStone.Player == "" {
				TempStone.Player = "Taken"
				TCount++
				grid[ypos-1][xpos] = TempStone
			}
		}
		if ypos < len(grid)-1 {
			TempStone := grid[ypos+1][xpos]
			if TempStone.Player == "" {
				TempStone.Player = "Taken"
				TCount++
				grid[ypos+1][xpos] = TempStone
			}
		}
	}
	for yindex := 0; yindex < len(grid); yindex = yindex + 1 {
		for xindex := 0; xindex < len(grid); xindex = xindex + 1 {
			TempStone := grid[yindex][xindex]
			if TempStone.Player == "Taken" {
				TempStone.Player = ""
				grid[yindex][xindex] = TempStone
			}
		}
	}
	return TCount
}

func makeMove(grid [][]Stones, player string, ypos int, xpos int, turn int) ([][]Stones, int) {
	PlacedStone := new(Stones)
	PlacedStone.Player = player
	PlacedStone.Turn = turn
	if isLegalMove(grid, player, ypos, xpos) == 1 {
		grid[ypos][xpos] = *PlacedStone
		for yindex := 0; yindex < len(grid); yindex++ {
			for xindex := 0; xindex < len(grid); xindex++ {
				if grid[ypos][xpos].Player != "" {
					//fmt.Println(grid)
					EmptyChain := make([]int, len(grid)*len(grid))
					StoneChain, ChainLen := FindStoneChain(grid, yindex, xindex, player, EmptyChain, 0)
					//fmt.Println(StoneChain, ChainLen)
					TerrorityCount := CountTerritory(grid, StoneChain, ChainLen)
					if TerrorityCount == 0 {
						//fmt.Println("Removing pieces")
						for index := 0; index < ChainLen*2; index = index + 2 {
							xremove := StoneChain[index]
							yremove := StoneChain[index]
							grid[yremove][xremove].Player = ""
							grid[yremove][xremove].Turn = 0
						}
					}
				}
			}
		}

		return grid, turn + 1
	}
	return grid, turn + 1 //Got an invalid move
}
