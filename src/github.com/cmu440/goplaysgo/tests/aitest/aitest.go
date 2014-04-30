package main

import (
	"fmt"
	"github.com/cmu440/goplaysgo/gogame"
	)

func TestRows(size int) bool {
	testBoard := gogame.MakeBoard(size)
	AnswerBoard := gogame.MakeBoard(len(testBoard.Grid))
	Turn := 1
	for yindex := 0; yindex < len(testBoard.Grid); yindex++ {
		//Place all the stones
		for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
			testBoard.MakeMove(gogame.Black, gogame.Move{yindex, xindex})
			AnswerBoard.Grid[yindex][xindex] = gogame.Stones{gogame.Black, Turn}
			Turn++
		}
		//Test all the stones
		for yindex2 := 0; yindex2 < len(testBoard.Grid); yindex2++ {
			for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
				if testBoard.Grid[yindex2][xindex] == AnswerBoard.Grid[yindex2][xindex] {
					testBoard.Grid[yindex2][xindex] = gogame.Stones{"", 0}
					AnswerBoard.Grid[yindex2][xindex] = gogame.Stones{"", 0}
				} else {
					//fmt.Println("Test Failed with piece at ypos:", yindex2, " xpos:", xindex)
					return false
				}
			}
		}
		testBoard.Turn = 1
		Turn = 1
	}
	return true
}

func TestCols(size int) bool {
	testBoard := gogame.MakeBoard(size)
	AnswerBoard := gogame.MakeBoard(len(testBoard.Grid))
	Turn := 1
	for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
		//Place all the stones
		for yindex := 0; yindex < len(testBoard.Grid); yindex++ {
			testBoard.MakeMove(gogame.Black, gogame.Move{yindex, xindex})
			AnswerBoard.Grid[yindex][xindex] = gogame.Stones{gogame.Black, Turn}
			Turn++
		}
		//Test all the stones
		for yindex2 := 0; yindex2 < len(testBoard.Grid); yindex2++ {
			for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
				if testBoard.Grid[yindex2][xindex] == AnswerBoard.Grid[yindex2][xindex] {
					testBoard.Grid[yindex2][xindex] = gogame.Stones{"", 0}
					AnswerBoard.Grid[yindex2][xindex] = gogame.Stones{"", 0}
				} else {
					//fmt.Println("Test Failed with piece at ypos:", yindex2, " xpos:", xindex)
					return false
				}
			}
		}
		testBoard.Turn = 1
		Turn = 1
	}
	return true
}

func TestSquare(size int) bool {
	testBoard := gogame.MakeBoard(size)
	AnswerBoard := gogame.MakeBoard(len(testBoard.Grid))
	Turn := 1
	//Top and bottom edge
	for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
		testBoard.MakeMove(gogame.Black, gogame.Move{0, xindex})
		AnswerBoard.Grid[0][xindex] = gogame.Stones{gogame.Black, Turn}
		Turn++
		testBoard.MakeMove(gogame.Black, gogame.Move{len(testBoard.Grid) - 1, xindex})
		AnswerBoard.Grid[len(testBoard.Grid)-1][xindex] = gogame.Stones{gogame.Black, Turn}
		Turn++
	}
	//Left and right edge
	//testBoard.printBoard()
	for yindex := 1; yindex < len(testBoard.Grid)-1; yindex++ {
		//fmt.Println("Yindex:",yindex) //Fails at Y equals 3
		testBoard.MakeMove(gogame.Black, gogame.Move{yindex, 0})
		AnswerBoard.Grid[yindex][0] = gogame.Stones{gogame.Black, Turn}
		Turn++
		testBoard.MakeMove(gogame.Black, gogame.Move{yindex, len(testBoard.Grid) - 1})
		AnswerBoard.Grid[yindex][len(testBoard.Grid)-1] = gogame.Stones{gogame.Black, Turn}
		Turn++
	}
	for yindex := 0; yindex < len(testBoard.Grid); yindex++ {
		for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
			if testBoard.Grid[yindex][xindex] == AnswerBoard.Grid[yindex][xindex] {
				testBoard.Grid[yindex][xindex] = gogame.Stones{"", 0}
				AnswerBoard.Grid[yindex][xindex] = gogame.Stones{"", 0}
			} else {
				for index := 0; index < len(testBoard.Grid); index = index + 1 {
					fmt.Println(testBoard.Grid[index])
				}
				return false
			}
		}
	}
	testBoard.Turn = 1
	return true
}

func TestInnerSquare(size int) bool {
	testBoard := gogame.MakeBoard(size)
	AnswerBoard := gogame.MakeBoard(len(testBoard.Grid))
	Turn := 1
	//Top and bottom edge
	for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
		testBoard.MakeMove(gogame.Black, gogame.Move{0, xindex})
		AnswerBoard.Grid[0][xindex] = gogame.Stones{gogame.Black, Turn}
		Turn++
		testBoard.MakeMove(gogame.Black, gogame.Move{len(testBoard.Grid) - 1, xindex})
		AnswerBoard.Grid[len(testBoard.Grid)-1][xindex] = gogame.Stones{gogame.Black, Turn}
		Turn++
	}
	//Left and right edge
	for yindex := 1; yindex < len(testBoard.Grid)-1; yindex++ {
		testBoard.MakeMove(gogame.Black,gogame.Move{yindex, 0})
		AnswerBoard.Grid[yindex][0] = gogame.Stones{gogame.Black, Turn}
		Turn++
		testBoard.MakeMove(gogame.Black, gogame.Move{yindex, len(testBoard.Grid) - 1})
		AnswerBoard.Grid[yindex][len(testBoard.Grid)-1] = gogame.Stones{gogame.Black, Turn}
		Turn++
	}

	//Test Inner Square
	testBoard.MakeMove(gogame.Black, gogame.Move{(len(testBoard.Grid) / 2), (len(testBoard.Grid) / 2)})
	testBoard.MakeMove(gogame.Black, gogame.Move{(len(testBoard.Grid) / 2), (len(testBoard.Grid) / 2) + 1})
	testBoard.MakeMove(gogame.Black, gogame.Move{(len(testBoard.Grid) / 2) + 1, (len(testBoard.Grid) / 2)})
	testBoard.MakeMove(gogame.Black, gogame.Move{(len(testBoard.Grid) / 2) + 1, (len(testBoard.Grid) / 2) + 1})

	AnswerBoard.Grid[(len(testBoard.Grid) / 2)][(len(testBoard.Grid) / 2)] = gogame.Stones{gogame.Black, Turn}
	AnswerBoard.Grid[(len(testBoard.Grid) / 2)][(len(testBoard.Grid)/2)+1] = gogame.Stones{gogame.Black, Turn + 1}
	AnswerBoard.Grid[(len(testBoard.Grid)/2)+1][(len(testBoard.Grid) / 2)] = gogame.Stones{gogame.Black, Turn + 2}
	AnswerBoard.Grid[(len(testBoard.Grid)/2)+1][(len(testBoard.Grid)/2)+1] = gogame.Stones{gogame.Black, Turn + 3}
	for yindex := 0; yindex < len(testBoard.Grid); yindex++ {
		for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
			if testBoard.Grid[yindex][xindex] == AnswerBoard.Grid[yindex][xindex] {
				testBoard.Grid[yindex][xindex] = gogame.Stones{"", 0}
				AnswerBoard.Grid[yindex][xindex] = gogame.Stones{"", 0}
			} else {
				//fmt.Println("Test Failed with piece at ypos:", yindex, " xpos:", xindex)
				return false
			}
		}
	}
	return true
}

func TestCapture() bool {
	testBoard := gogame.MakeBoard(5)
	AnswerBoard := gogame.MakeBoard(len(testBoard.Grid))
	testBoard.MakeMove(gogame.Black, gogame.Move{1, 2})
	testBoard.MakeMove(gogame.Black, gogame.Move{2, 1})
	testBoard.MakeMove(gogame.Black, gogame.Move{2, 3})
	testBoard.MakeMove(gogame.White, gogame.Move{2, 2})
	testBoard.MakeMove(gogame.Black, gogame.Move{3, 2})

	AnswerBoard.Grid[1][2] = gogame.Stones{gogame.Black, 1}
	AnswerBoard.Grid[2][1] = gogame.Stones{gogame.Black, 2}
	AnswerBoard.Grid[2][3] = gogame.Stones{gogame.Black, 3}
	AnswerBoard.Grid[3][2] = gogame.Stones{gogame.Black, 5}

	for yindex := 0; yindex < len(testBoard.Grid); yindex++{
		for xindex := 0; xindex < len(testBoard.Grid) ; xindex++{
			if testBoard.Grid[yindex][xindex] != AnswerBoard.Grid[yindex][xindex] {
				for index := 0; index < len(testBoard.Grid); index = index + 1 {
					fmt.Println(testBoard.Grid[index])
				}
				return false
			}
		}
	}
	return true
}

func main(){
	var Test bool;
	var Test1 bool;
	Sizes := make([]int,3)
	Sizes[0] = 5
	Sizes[1] = 9
	Sizes[2] = 13
	for index := 0; index < len(Sizes); index++ {
		Test = TestRows(Sizes[index])
		Test1 = TestCols(Sizes[index])
		if Test && Test1 {
			fmt.Println("Passed test at size:",Sizes[index])
		} else {
			fmt.Println("Failed test at size:",Sizes[index])
			return
		}
	}
	
	Test = TestCapture() 
	if Test {
		fmt.Println("Make move is capturing correctly")
	} else {
		fmt.Println("Make move is NOT capturing correctly")
	}
}
