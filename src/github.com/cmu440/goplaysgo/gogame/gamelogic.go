package main

import "fmt"

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
	Bd.Turn = 0
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

func FindPosChain(chain []int, xpos int, ypos int) int {
	for index := 0; index < len(chain); index += 2 {
		xindex := chain[index]
		yindex := chain[index+1]
		if (xpos == xindex) && (ypos == yindex) {
			return 0
		}
	}
	return 1
}

func FindStoneChain(grid [][]Stones, xpos int, ypos int, player string, chain []int, count int) ([]int, int) {
	fmt.Println("Called FindStoneChain at:", xpos, " ", ypos)
	if xpos > 0 {
		TempStoneLeft := grid[xpos-1][ypos]
		fmt.Println("TempStoneLeft:", TempStoneLeft)
		if (TempStoneLeft.Player == player) && (FindPosChain(chain, xpos, ypos) == 1) {
			fmt.Println("Turned left")
			chain[count] = xpos - 1
			count++
			chain[count] = ypos
			count++
			chain, count = FindStoneChain(grid, xpos-1, ypos, player, chain, count)
		}
	}
	if len(grid)-1 > xpos {
		TempStoneRight := grid[xpos+1][ypos]
		fmt.Println("TempStoneRight:", TempStoneRight)
		if (TempStoneRight.Player == player) &&(FindPosChain(chain, xpos, ypos) == 1) {
			fmt.Println("Turned right")
			chain[count] = xpos + 1
			count++
			chain[count] = ypos
			count++
			chain, count = FindStoneChain(grid, xpos+1, ypos, player, chain, count)
		}
	}
	if ypos > 0 {
		TempStoneUp := grid[xpos][ypos-1]
		fmt.Println("TempStoneUp:", TempStoneUp)
		if (TempStoneUp.Player == player) && (FindPosChain(chain, xpos, ypos) == 1) {
			fmt.Println("Turned Up")
			chain[count] = xpos
			count++
			chain[count] = ypos - 1
			count++
			chain, count = FindStoneChain(grid, xpos, ypos-1, player, chain, count)
		}
	}
	if len(grid)-1 > ypos {
		TempStoneDown := grid[xpos][ypos+1]
		fmt.Println("TempStoneDown:", TempStoneDown)
		if (TempStoneDown.Player == player) && (FindPosChain(chain, xpos, ypos) == 1) {
			fmt.Println("Turned Down")
			chain[count] = xpos
			count++
			chain[count] = ypos + 1
			count++
			chain, count = FindStoneChain(grid, xpos, ypos+1, player, chain, count)
		}
	}
	chain[count] = xpos
	count++
	chain[count] = ypos
	count++
	return chain, count
}

/*
func removePieces (grid [][]Stones,turn int) [][]Stones {
	RemoveList := make([]int,turn)
	RemoveCount := 0
	for yindex := 0; yindex < len(grid); yindex++ {
	    for xindex :=0; xindex < len(grid); xindex++ {
	        TestStone := grid[yindex][xindex]
		if (TestStone.Player != "")


*/

func makeMove(grid [][]Stones, player string, xpos int, ypos int, turn int) [][]Stones {
	PlacedStone := new(Stones)
	PlacedStone.Player = player
	PlacedStone.Turn = turn
	if isLegalMove(grid, player, xpos, ypos) == 1 {
		grid[ypos][xpos] = *PlacedStone
		return grid
	}
	return grid //Got an invalid move
}
