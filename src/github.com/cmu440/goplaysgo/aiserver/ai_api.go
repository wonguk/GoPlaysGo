package ai

type AI interface {
	// NextMove should be implemented by the AI coders, where it
	// returns the next move the ai wants to play given the current
	// state of the game.
	// TODO: Decide if we want to specify a timeout
	// TODO: Decide what will happen if errors are caught
	NextMove(gogame.Board, gogame.Player) gogame.Move
}
