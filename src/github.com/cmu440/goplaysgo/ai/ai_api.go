package ai

import "github.com/cmu440/goplaysgo/gogame"

// AI is the interface AI coders should implement
type AI interface {
	// NextMove should be implemented by the AI coders, where it
	// returns the next move the ai wants to play given the current
	// state of the game.
	NextMove(gogame.Board, gogame.Player) gogame.Move
}
