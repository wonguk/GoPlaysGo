package aiserver

import (
	"github.com/cmu440/goplaysgo/gogame"
	"github.com/cmu440/goplaysgo/rpc/airpc"
)

type aiServer struct {
}

func NewAIServer(port int) (AIServer, error) {
	return nil, errors.New("not implemented")
}

func (as *aiServer) NextMove(args *airpc.NextMoveArgs, reply *airpc.NextMoveReply) error {
	return errors.New("not implemented")
}

func (as *aiServer) CheckGame(args *airpc.CheckArgs, reply *airpc.CheckReply) error {
	return errors.New("not implemented")
}

func (as *aiServer) InitGame(args *airpc.InitGameArgs, reply *airpc.InitGameReply) error {
	return errors.New("not implemented")
}

func (as *aiServer) StartGame(args *airpc.StartGameArgs, reply *airpc.StartGameReply) error {
	return errors.New("not implemented")
}
