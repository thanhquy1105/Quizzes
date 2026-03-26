package trace

import "btaskee-quiz/pkg/wkhttp"

type ClusterKind int

const (
	ClusterKindUnknown ClusterKind = iota

	ClusterKindSlot

	ClusterKindChannel

	ClusterKindConfig

	ClusterKindOther
)

type IMetrics interface {
	System() ISystemMetrics

	App() IAppMetrics

	Cluster() IClusterMetrics

	DB() IDBMetrics

	Route(r *wkhttp.WKHttp)
}

type ISystemMetrics interface {
	IntranetIncomingAdd(v int64)

	IntranetOutgoingAdd(v int64)

	ExtranetIncomingAdd(v int64)

	ExtranetOutgoingAdd(v int64)

	CPUUsageAdd(v float64)

	MemoryUsageAdd(v float64)

	DiskIOReadCountAdd(v int64)

	DiskIOWriteCountAdd(v int64)
}

type IDBMetrics interface {
	CompactTotalCountSet(shardId uint32, v int64)

	CompactDefaultCountSet(shardId uint32, v int64)

	CompactDeleteOnlyCountSet(shardId uint32, v int64)

	CompactElisionOnlyCountSet(shardId uint32, v int64)

	CompactMoveCountSet(shardId uint32, v int64)

	CompactReadCountSet(shardId uint32, v int64)

	CompactRewriteCountSet(shardId uint32, v int64)

	CompactMultiLevelCount(shardId uint32, v int64)

	CompactEstimatedDebtSet(shardId uint32, v int64)

	CompactInProgressBytesSet(shardId uint32, v int64)

	CompactNumInProgressSet(shardId uint32, v int64)

	CompactMarkedFilesSet(shardId uint32, v int64)

	FlushCountAdd(shardId uint32, v int64)

	FlushBytesAdd(shardId uint32, v int64)

	FlushNumInProgressAdd(shardId uint32, v int64)

	FlushAsIngestCountAdd(shardId uint32, v int64)

	FlushAsIngestTableCountAdd(shardId uint32, v int64)

	FlushAsIngestBytesAdd(shardId uint32, v int64)

	MemTableSizeSet(shardId uint32, v int64)
	MemTableCountSet(shardId uint32, v int64)

	MemTableZombieSizeSet(shardId uint32, v int64)

	MemTableZombieCountSet(shardId uint32, v int64)

	SnapshotsCountSet(shardId uint32, v int64)

	TableCacheSizeSet(shardId uint32, v int64)

	TableCacheCountSet(shardId uint32, v int64)

	TableItersCountSet(shardId uint32, v int64)

	WALFilesCountSet(shardId uint32, v int64)

	WALSizeSet(shardId uint32, v int64)

	WALPhysicalSizeSet(shardId uint32, v int64)

	WALObsoleteFilesCountSet(shardId uint32, v int64)

	WALObsoletePhysicalSizeSet(shardId uint32, v int64)

	WALBytesInSet(shardId uint32, v int64)

	WALBytesWrittenSet(shardId uint32, v int64)

	LogWriterBytesSet(shardId uint32, v int64)

	DiskSpaceUsageSet(shardId uint32, v int64)

	LevelNumFilesSet(shardId uint32, v int64)
	LevelFileSizeSet(shardId uint32, v int64)
	LevelCompactScoreSet(shardId uint32, v int64)
	LevelBytesInSet(shardId uint32, v int64)
	LevelBytesIngestedSet(shardId uint32, v int64)
	LevelBytesMovedSet(shardId uint32, v int64)
	LevelBytesReadSet(shardId uint32, v int64)
	LevelBytesCompactedSet(shardId uint32, v int64)
	LevelBytesFlushedSet(shardId uint32, v int64)
	LevelTablesCompactedSet(shardId uint32, v int64)
	LevelTablesFlushedSet(shardId uint32, v int64)
	LevelTablesIngestedSet(shardId uint32, v int64)
	LevelTablesMovedSet(shardId uint32, v int64)

	MessageAppendBatchCountAdd(v int64)

	SetAdd(v int64)
	DeleteAdd(v int64)
	DeleteRangeAdd(v int64)
	CommitAdd(v int64)

	AddAllowlistAdd(v int64)
	GetAllowlistAdd(v int64)
	HasAllowlistAdd(v int64)
	ExistAllowlistAdd(v int64)
	RemoveAllowlistAdd(v int64)
	RemoveAllAllowlistAdd(v int64)

	SaveChannelClusterConfigAdd(v int64)
	SaveChannelClusterConfigsAdd(v int64)
	GetChannelClusterConfigAdd(v int64)
	GetChannelClusterConfigVersionAdd(v int64)
	GetChannelClusterConfigsAdd(v int64)
	SearchChannelClusterConfigAdd(v int64)
	GetChannelClusterConfigCountWithSlotIdAdd(v int64)
	GetChannelClusterConfigWithSlotIdAdd(v int64)

	AddChannelAdd(v int64)
	UpdateChannelAdd(v int64)
	GetChannelAdd(v int64)
	SearchChannelsAdd(v int64)
	ExistChannelAdd(v int64)
	UpdateChannelAppliedIndexAdd(v int64)
	GetChannelAppliedIndexAdd(v int64)
	DeleteChannelAdd(v int64)

	AddOrUpdateConversationsAddWithUser(v int64)
	AddOrUpdateConversationsAdd(v int64)
	GetConversationsAdd(v int64)
	GetConversationsByTypeAdd(v int64)
	GetLastConversationsAdd(v int64)
	GetConversationAdd(v int64)
	ExistConversationAdd(v int64)
	DeleteConversationAdd(v int64)
	DeleteConversationsAdd(v int64)
	SearchConversationAdd(v int64)
	AddDenylistAdd(v int64)
	GetDenylistAdd(v int64)
	ExistDenylistAdd(v int64)
	RemoveDenylistAdd(v int64)
	RemoveAllDenylistAdd(v int64)

	GetDeviceAdd(v int64)
	GetDevicesAdd(v int64)
	GetDeviceCountAdd(v int64)
	AddDeviceAdd(v int64)
	UpdateDeviceAdd(v int64)
	SearchDeviceAdd(v int64)

	AppendMessageOfNotifyQueueAdd(v int64)
	GetMessagesOfNotifyQueueAdd(v int64)
	RemoveMessagesOfNotifyQueueAdd(v int64)

	AppendMessagesAdd(v int64)
	AppendMessagesBatchAdd(v int64)
	GetMessageAdd(v int64)
	LoadPrevRangeMsgsAdd(v int64)
	LoadNextRangeMsgsAdd(v int64)
	LoadMsgAdd(v int64)
	LoadLastMsgsAdd(v int64)
	LoadLastMsgsWithEndAdd(v int64)
	LoadNextRangeMsgsForSizeAdd(v int64)
	TruncateLogToAdd(v int64)
	GetChannelLastMessageSeqAdd(v int64)
	SetChannelLastMessageSeqAdd(v int64)
	SearchMessagesAdd(v int64)

	AddSubscribersAdd(v int64)
	GetSubscribersAdd(v int64)
	RemoveSubscribersAdd(v int64)
	ExistSubscriberAdd(v int64)
	RemoveAllSubscriberAdd(v int64)

	AddSystemUidsAdd(v int64)
	RemoveSystemUidsAdd(v int64)
	GetSystemUidsAdd(v int64)

	GetUserAdd(v int64)
	ExistUserAdd(v int64)
	SearchUserAdd(v int64)
	AddUserAdd(v int64)
	UpdateUserAdd(v int64)

	SetLeaderTermStartIndexAdd(v int64)
	LeaderLastTermAdd(v int64)
	LeaderTermStartIndexAdd(v int64)
	LeaderLastTermGreaterThanAdd(v int64)
	DeleteLeaderTermStartIndexGreaterThanTermAdd(v int64)
}

type IAppMetrics interface {
	ConnCountAdd(v int64)

	OnlineUserCountAdd(v int64)

	OnlineUserCountSet(v int64)

	OnlineDeviceCountAdd(v int64)

	OnlineDeviceCountSet(v int64)

	MessageLatencyOb(v int64)

	PingBytesAdd(v int64)
	PingBytes() int64

	PingCountAdd(v int64)
	PingCount() int64

	PongBytesAdd(v int64)
	PongBytes() int64

	PongCountAdd(v int64)
	PongCount() int64

	SendPacketBytesAdd(v int64)
	SendPacketBytes() int64

	SendPacketCountAdd(v int64)
	SendPacketCount() int64

	SendackPacketBytesAdd(v int64)
	SendackPacketBytes() int64

	SendackPacketCountAdd(v int64)
	SendackPacketCount() int64

	RecvPacketBytesAdd(v int64)
	RecvPacketBytes() int64

	RecvPacketCountAdd(v int64)
	RecvPacketCount() int64

	RecvackPacketBytesAdd(v int64)
	RecvackPacketBytes() int64

	RecvackPacketCountAdd(v int64)
	RecvackPacketCount() int64

	ConnPacketBytesAdd(v int64)
	ConnPacketBytes() int64

	ConnPacketCountAdd(v int64)
	ConnPacketCount() int64

	ConnackPacketBytesAdd(v int64)
	ConnackPacketBytes() int64

	ConnackPacketCountAdd(v int64)
	ConnackPacketCount() int64
}

type IClusterMetrics interface {
	MessageIncomingBytesAdd(kind ClusterKind, v int64)

	MessageOutgoingBytesAdd(kind ClusterKind, v int64)

	MessageIncomingCountAdd(kind ClusterKind, v int64)

	MessageOutgoingCountAdd(kind ClusterKind, v int64)

	MessageConcurrencyAdd(v int64)

	SendPacketIncomingBytesAdd(v int64)

	SendPacketOutgoingBytesAdd(v int64)

	SendPacketIncomingCountAdd(v int64)

	SendPacketOutgoingCountAdd(v int64)

	RecvPacketIncomingBytesAdd(v int64)

	RecvPacketOutgoingBytesAdd(v int64)

	RecvPacketIncomingCountAdd(v int64)

	RecvPacketOutgoingCountAdd(v int64)

	MsgSyncIncomingBytesAdd(kind ClusterKind, v int64)

	MsgSyncIncomingCountAdd(kind ClusterKind, v int64)

	MsgSyncOutgoingBytesAdd(kind ClusterKind, v int64)

	MsgSyncOutgoingCountAdd(kind ClusterKind, v int64)

	MsgSyncRespIncomingBytesAdd(kind ClusterKind, v int64)

	MsgSyncRespIncomingCountAdd(kind ClusterKind, v int64)

	MsgSyncRespOutgoingBytesAdd(kind ClusterKind, v int64)

	MsgSyncRespOutgoingCountAdd(kind ClusterKind, v int64)

	MsgClusterPingIncomingBytesAdd(kind ClusterKind, v int64)

	MsgClusterPingIncomingCountAdd(kind ClusterKind, v int64)

	MsgClusterPingOutgoingBytesAdd(kind ClusterKind, v int64)

	MsgClusterPingOutgoingCountAdd(kind ClusterKind, v int64)

	MsgClusterPongIncomingBytesAdd(kind ClusterKind, v int64)

	MsgClusterPongIncomingCountAdd(kind ClusterKind, v int64)

	MsgClusterPongOutgoingBytesAdd(kind ClusterKind, v int64)

	MsgClusterPongOutgoingCountAdd(kind ClusterKind, v int64)

	LogIncomingBytesAdd(kind ClusterKind, v int64)

	LogIncomingCountAdd(kind ClusterKind, v int64)

	LogOutgoingBytesAdd(kind ClusterKind, v int64)

	LogOutgoingCountAdd(kind ClusterKind, v int64)

	MsgLeaderTermStartIndexReqIncomingBytesAdd(kind ClusterKind, v int64)

	MsgLeaderTermStartIndexReqIncomingCountAdd(kind ClusterKind, v int64)

	MsgLeaderTermStartIndexReqOutgoingBytesAdd(kind ClusterKind, v int64)

	MsgLeaderTermStartIndexReqOutgoingCountAdd(kind ClusterKind, v int64)

	MsgLeaderTermStartIndexRespIncomingBytesAdd(kind ClusterKind, v int64)

	MsgLeaderTermStartIndexRespIncomingCountAdd(kind ClusterKind, v int64)

	MsgLeaderTermStartIndexRespOutgoingBytesAdd(kind ClusterKind, v int64)

	MsgLeaderTermStartIndexRespOutgoingCountAdd(kind ClusterKind, v int64)

	ForwardProposeBytesAdd(v int64)

	ForwardProposeCountAdd(v int64)

	ForwardProposeRespBytesAdd(v int64)

	ForwardProposeRespCountAdd(v int64)

	ForwardConnPingBytesAdd(v int64)

	ForwardConnPingCountAdd(v int64)

	ForwardConnPongBytesAdd(v int64)

	ForwardConnPongCountAdd(v int64)

	ChannelActiveCountAdd(v int64)

	ChannelElectionCountAdd(v int64)

	ChannelElectionSuccessCountAdd(v int64)

	ChannelElectionFailCountAdd(v int64)

	SlotElectionCountAdd(v int64)

	SlotElectionSuccessCountAdd(v int64)

	SlotElectionFailCountAdd(v int64)

	ProposeLatencyAdd(kind ClusterKind, v int64)

	ProposeFailedCountAdd(kind ClusterKind, v int64)

	ObserverNodeRequesting(f func() int64)

	ObserverNodeSending(f func() int64)
}
