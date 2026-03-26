package raftgroup

import "btaskee-quiz/pkg/raft/types"

type Event struct {
	RaftKey string
	types.Event
	WaitC chan error
}
