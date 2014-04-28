package paxosrpc

type Status int

const (
	OK Status = iota + 1
	NotReady
	Reject
)

const (
	PaxosTimeout = 10
)

type PrepareArgs struct {
	N int
}

type PrepareReply struct {
	Status Status
}

type AcceptArgs struct {
}

type AcceptReply struct {
	Status Status
}

type CommitArgs struct {
}

type CommitReply struct {
	Status Status
}
