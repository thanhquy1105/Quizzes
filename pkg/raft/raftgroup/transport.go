package raftgroup

import "btaskee-quiz/pkg/raft/types"

type ITransport interface {
	// Send 发送事件
	Send(key string, event types.Event)
}
