package raft

import "btaskee-quiz/pkg/raft/types"

type stepReq struct {
	event types.Event
	resp  chan error
}
