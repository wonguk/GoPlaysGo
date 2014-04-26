package mainserver

import (
	"github.com/cmu440/goplaysgo/rpc/mainrpc"
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
}

type statsMaster struct {
	reqChan    chan statsRequest
	allReqChan chan allStatsRequest

	initChan chan initRequest
	addChan  chan mainrpc.GameResult

	stats map[string]mainrpc.Stats
}

func (sm *statsMaster) startStatsMaster() {
	for {
		select {
		case req := <-sm.reqChan:
			s := sm.stats[req.name]
			req.retChan <- s

		case req := <-sm.allReqChan:
			stats := make([]mainrpc.Stats, len(sm.stats))

			for _, s := range sm.stats {
				stats = append(stats, s)
			}

			req.retChan <- stats

		case init := <-sm.initChan:
			if _, ok := sm.stats[init.name]; ok {
				continue
			}

			//TODO PAXOS

			sm.stats[init.name] = initStats(init.name, init.hostport)

		case res := <-sm.addChan:
			if _, ok := sm.stats[res.Player1]; !ok {
				continue
			}
			if _, ok := sm.stats[res.Player2]; !ok {
				continue
			}

			//TODO PAXOS

			switch {
			case res.Points1 < res.Points2:
				sm.stats[res.Player1] = updateLoss(sm.stats[res.Player1], res)
				sm.stats[res.Player2] = updateWin(sm.stats[res.Player2], res)

			case res.Points2 < res.Points1:
				sm.stats[res.Player1] = updateWin(sm.stats[res.Player1], res)
				sm.stats[res.Player2] = updateLoss(sm.stats[res.Player2], res)

			case res.Points1 == res.Points2:
				sm.stats[res.Player1] = updateDraw(sm.stats[res.Player1], res)
				sm.stats[res.Player1] = updateDraw(sm.stats[res.Player1], res)
			}
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
