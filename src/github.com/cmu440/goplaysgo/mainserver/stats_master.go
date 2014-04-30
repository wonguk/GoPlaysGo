package mainserver

import (
	"container/list"

	"github.com/cmu440/goplaysgo/rpc/mainrpc"
	"github.com/cmu440/goplaysgo/rpc/paxosrpc"
)

type statsRequest struct {
	name    string
	retChan chan mainrpc.Stats
}

type allStatsRequest struct {
	retChan chan mainrpc.Standings
}

type initRequest struct {
	name     string
	hostport string

	done chan bool
}

type resultRequest struct {
	result  mainrpc.GameResult
	retChan chan mainrpc.Status
}

type statsMaster struct {
	reqChan    chan statsRequest
	allReqChan chan allStatsRequest

	initChan chan initRequest
	//addChan  chan mainrpc.GameResult
	resultChan chan resultRequest

	commitChan chan paxosrpc.Command

	stats    map[string]mainrpc.Stats
	toRun    map[string]chan bool
	toReturn *list.List
}

func (sm *statsMaster) startStatsMaster(cmdChan chan paxosrpc.Command,
	aiChan chan *newAIReq) {
	for {
		select {
		case req := <-sm.reqChan:
			s := sm.stats[req.name]
			req.retChan <- s

		case req := <-sm.allReqChan:
			stats := make([]mainrpc.Stats, len(sm.stats))

			i := 0
			for _, s := range sm.stats {
				stats[i] = s
				i++
			}

			req.retChan <- stats

		case init := <-sm.initChan:
			if _, ok := sm.stats[init.name]; ok {
				continue
			}

			sm.toRun[init.name] = init.done

			//PAXOS
			cmd := paxosrpc.Command{
				Type:          paxosrpc.Init,
				CommandNumber: -1,
				Player:        init.name,
				Hostport:      init.hostport,
			}
			cmdChan <- cmd
			/*
				case res := <-sm.addChan:
					//PAXOS
					cmd := paxosrpc.Command{
						Type:          paxosrpc.Update,
						CommandNumber: -1,
						GameResult:    res,
					}
					cmdChan <- cmd
			*/
		case req := <-sm.resultChan:
			//PAXOS
			cmd := paxosrpc.Command{
				Type:          paxosrpc.Update,
				CommandNumber: -1,
				GameResult:    req.result,
			}
			cmdChan <- cmd

			//Save Request
			sm.toReturn.PushBack(req)

		case cmd := <-sm.commitChan:
			switch cmd.Type {
			case paxosrpc.Init:
				_, ok := sm.stats[cmd.Player]

				if ok {
					if c, ok := sm.toRun[cmd.Player]; ok {
						c <- false
						delete(sm.toRun, cmd.Player)
					}
				} else {
					sm.stats[cmd.Player] = initStats(cmd.Player, cmd.Hostport)

					if c, ok := sm.toRun[cmd.Player]; ok {
						c <- true
						delete(sm.toRun, cmd.Player)
					}
					req := new(newAIReq)
					req.name = cmd.Player
					req.manage = false
					req.hostport = cmd.Hostport

					aiChan <- req
				}

			case paxosrpc.Update:
				res := cmd.GameResult

				sm.handleResult(res)

				if _, ok := sm.stats[res.Player1]; !ok {
					continue
				}
				if _, ok := sm.stats[res.Player2]; !ok {
					continue
				}

				switch {
				case res.Points1 < res.Points2:
					sm.stats[res.Player1] = updateLoss(sm.stats[res.Player1], res)
					sm.stats[res.Player2] = updateWin(sm.stats[res.Player2], res)

				case res.Points2 < res.Points1:
					sm.stats[res.Player1] = updateWin(sm.stats[res.Player1], res)
					sm.stats[res.Player2] = updateLoss(sm.stats[res.Player2], res)

				case res.Points1 == res.Points2:
					sm.stats[res.Player1] = updateDraw(sm.stats[res.Player1], res)
					sm.stats[res.Player2] = updateDraw(sm.stats[res.Player2], res)
				}
			default:
			}
		}

	}
}

func (sm *statsMaster) handleResult(res mainrpc.GameResult) {
	for e := sm.toReturn.Front(); e != nil; e = e.Next() {
		req := e.Value.(resultRequest)
		r := req.result

		if r.Player1 == res.Player1 && r.Player2 == res.Player2 {
			sm.toReturn.Remove(e)

			req.retChan <- mainrpc.OK
			return
		}
	}
}

func updateWin(s mainrpc.Stats, res mainrpc.GameResult) mainrpc.Stats {
	s.Wins++
	s.GameResults = append(s.GameResults, res)

	return s
}

func updateLoss(s mainrpc.Stats, res mainrpc.GameResult) mainrpc.Stats {
	s.Losses++
	s.GameResults = append(s.GameResults, res)

	return s
}

func updateDraw(s mainrpc.Stats, res mainrpc.GameResult) mainrpc.Stats {
	s.Draws++
	s.GameResults = append(s.GameResults, res)

	return s
}

func initStats(name string, hostport string) mainrpc.Stats {
	s := mainrpc.Stats{
		Name:        name,
		Hostport:    hostport,
		Wins:        0,
		Losses:      0,
		Draws:       0,
		GameResults: []mainrpc.GameResult{},
	}

	return s
}
