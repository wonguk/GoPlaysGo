package mainrpc

type Status int

const (
	OK Status = iota + 1
	NotReady
	WrongServer
	AIExists
)

//TODO: Timing constraints for main server?
const ()

type Node struct {
	hostname string
	client   *rpc.Client
}

type GameResult struct {
	player1 string
	player2 string

	points1 int
	points2 int
}

type Stats struct {
	Name        string
	Hostport    string
	Wins        int
	Losses      int
	Draws       int
	GameResults []GameResult
}

type Standings []Stats

type RegisterArgs struct {
	hostname string
}

type RegisterReply struct {
	Status  Status
	Servers []string
}

type RegisterRefArgs struct {
	hostname string
}

type RegisterRefReply struct {
	Status Status
}

type SubmitAIArgs struct {
	name string
	code []byte
}

type SubmitAIReply struct {
	Status Status
}

type GetStandingsArgs struct {
}

type GetStandingReply struct {
	Status    Status
	Standings Standings
}
