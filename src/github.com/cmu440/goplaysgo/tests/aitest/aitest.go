package aitest

import "fmt"

func TestRows(testBoard *Board) bool {
	AnswerBoard := makeBoard(len(testBoard.Grid))
	Turn := 1
	for yindex := 0; yindex < len(testBoard.Grid); yindex++ {
		//Place all the stones
		for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
			testBoard.makeMove(Black, Move{yindex, xindex})
			AnswerBoard.Grid[yindex][xindex] = Stones{Black, Turn}
			Turn++
		}
		//Test all the stones
		for yindex2 := 0; yindex2 < len(testBoard.Grid); yindex2++ {
			for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
				if testBoard.Grid[yindex2][xindex] == AnswerBoard.Grid[yindex2][xindex] {
					testBoard.Grid[yindex2][xindex] = Stones{"", 0}
					AnswerBoard.Grid[yindex2][xindex] = Stones{"", 0}
				} else {
					fmt.Println("Test Failed with piece at ypos:", yindex2, " xpos:", xindex)
					return false
				}
			}
		}
		testBoard.Turn = 1
		Turn = 1
	}
	return true
}

func TestCols(testBoard *Board) bool {
	AnswerBoard := makeBoard(len(testBoard.Grid))
	Turn := 1
	for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
		//Place all the stones
		for yindex := 0; yindex < len(testBoard.Grid); yindex++ {
			testBoard.makeMove(Black, Move{yindex, xindex})
			AnswerBoard.Grid[yindex][xindex] = Stones{Black, Turn}
			Turn++
		}
		//Test all the stones
		for yindex2 := 0; yindex2 < len(testBoard.Grid); yindex2++ {
			for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
				if testBoard.Grid[yindex2][xindex] == AnswerBoard.Grid[yindex2][xindex] {
					testBoard.Grid[yindex2][xindex] = Stones{"", 0}
					AnswerBoard.Grid[yindex2][xindex] = Stones{"", 0}
				} else {
					fmt.Println("Test Failed with piece at ypos:", yindex2, " xpos:", xindex)
					return false
				}
			}
		}
		testBoard.Turn = 1
		Turn = 1
	}
	return true
}

func TestSquare(testBoard *Board) bool {
	AnswerBoard := makeBoard(len(testBoard.Grid))
	Turn := 1
	//Top and bottom edge
	for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
		testBoard.makeMove(Black, Move{0, xindex})
		AnswerBoard.Grid[0][xindex] = Stones{Black, Turn}
		Turn++
		testBoard.makeMove(Black, Move{len(testBoard.Grid) - 1, xindex})
		AnswerBoard.Grid[len(testBoard.Grid)-1][xindex] = Stones{Black, Turn}
		Turn++
	}
	//Left and right edge
	//testBoard.printBoard()
	for yindex := 1; yindex < len(testBoard.Grid)-1; yindex++ {
		//fmt.Println("Yindex:",yindex) //Fails at Y equals 3
		testBoard.makeMove(Black, Move{yindex, 0})
		AnswerBoard.Grid[yindex][0] = Stones{Black, Turn}
		Turn++
		testBoard.makeMove(Black, Move{yindex, len(testBoard.Grid) - 1})
		AnswerBoard.Grid[yindex][len(testBoard.Grid)-1] = Stones{Black, Turn}
		Turn++
	}
	for yindex := 0; yindex < len(testBoard.Grid); yindex++ {
		for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
			if testBoard.Grid[yindex][xindex] == AnswerBoard.Grid[yindex][xindex] {
				testBoard.Grid[yindex][xindex] = Stones{"", 0}
				AnswerBoard.Grid[yindex][xindex] = Stones{"", 0}
			} else {
				fmt.Println("Test Failed with piece at ypos:", yindex, " xpos:", xindex)
				return false
			}
		}
	}
	testBoard.Turn = 1
	return true
}

func TestInnerSquare(testBoard *Board) bool {
	AnswerBoard := makeBoard(len(testBoard.Grid))
	Turn := 1
	//Top and bottom edge
	for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
		testBoard.makeMove(Black, Move{0, xindex})
		AnswerBoard.Grid[0][xindex] = Stones{Black, Turn}
		Turn++
		testBoard.makeMove(Black, Move{len(testBoard.Grid) - 1, xindex})
		AnswerBoard.Grid[len(testBoard.Grid)-1][xindex] = Stones{Black, Turn}
		Turn++
	}
	//Left and right edge
	for yindex := 1; yindex < len(testBoard.Grid)-1; yindex++ {
		testBoard.makeMove(Black, Move{yindex, 0})
		AnswerBoard.Grid[yindex][0] = Stones{Black, Turn}
		Turn++
		testBoard.makeMove(Black, Move{yindex, len(testBoard.Grid) - 1})
		AnswerBoard.Grid[yindex][len(testBoard.Grid)-1] = Stones{Black, Turn}
		Turn++
	}

	//Test Inner Square
	testBoard.makeMove(Black, Move{(len(testBoard.Grid) / 2), (len(testBoard.Grid) / 2)})
	testBoard.makeMove(Black, Move{(len(testBoard.Grid) / 2), (len(testBoard.Grid) / 2) + 1})
	testBoard.makeMove(Black, Move{(len(testBoard.Grid) / 2) + 1, (len(testBoard.Grid) / 2)})
	testBoard.makeMove(Black, Move{(len(testBoard.Grid) / 2) + 1, (len(testBoard.Grid) / 2) + 1})

	AnswerBoard.Grid[(len(testBoard.Grid) / 2)][(len(testBoard.Grid) / 2)] = Stones{Black, Turn}
	AnswerBoard.Grid[(len(testBoard.Grid) / 2)][(len(testBoard.Grid)/2)+1] = Stones{Black, Turn + 1}
	AnswerBoard.Grid[(len(testBoard.Grid)/2)+1][(len(testBoard.Grid) / 2)] = Stones{Black, Turn + 2}
	AnswerBoard.Grid[(len(testBoard.Grid)/2)+1][(len(testBoard.Grid)/2)+1] = Stones{Black, Turn + 3}
	for yindex := 0; yindex < len(testBoard.Grid); yindex++ {
		for xindex := 0; xindex < len(testBoard.Grid); xindex++ {
			if testBoard.Grid[yindex][xindex] == AnswerBoard.Grid[yindex][xindex] {
				testBoard.Grid[yindex][xindex] = Stones{"", 0}
				AnswerBoard.Grid[yindex][xindex] = Stones{"", 0}
			} else {
				fmt.Println("Test Failed with piece at ypos:", yindex, " xpos:", xindex)
				return false
			}
		}
	}
	return true
}
