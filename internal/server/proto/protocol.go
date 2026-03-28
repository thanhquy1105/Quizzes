package proto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"btaskee-quiz/pkg/log"

	"go.uber.org/zap"
)

type MsgType uint8

const (
	Unknown MsgType = iota
	MsgTypeConnect
	MsgTypeConnack
	MsgTypeRequest
	MsgTypeResp
	MsgTypeHeartbeat
	MsgTypeMessage
	MsgTypeBatchMessage
)

const (
	MsgTypeLength    = 1
	MsgContentLength = 4
)

var (
	MagicNumberStart       = []byte("BTASKEE")
	MagicNumberStartLength = len(MagicNumberStart)
)

func (m MsgType) Uint8() uint8 {
	return uint8(m)
}

func (m MsgType) String() string {
	switch m {
	case MsgTypeConnect:
		return "MsgTypeConnect"
	case MsgTypeConnack:
		return "MsgTypeConnack"
	case MsgTypeRequest:
		return "MsgTypeRequest"
	case MsgTypeResp:
		return "MsgTypeResp"
	case MsgTypeHeartbeat:
		return "MsgTypeHeartbeat"
	case MsgTypeMessage:
		return "MsgTypeMessage"
	case MsgTypeBatchMessage:
		return "MsgTypeBatchMessage"
	default:
		return fmt.Sprintf("Unknown MsgType %d", m)
	}
}

type Reader interface {
	InboundBuffered() int
	Peek(n int) ([]byte, error)
	Discard(n int) (int, error)
}

type Protocol interface {
	Decode(c Reader) ([]byte, MsgType, int, error)
	Encode(data []byte, msgType MsgType) ([]byte, error)
}

type DefaultProto struct {
	log.Log
}

func New() *DefaultProto {

	return &DefaultProto{
		Log: log.NewBLog("server.proto"),
	}
}

func (d *DefaultProto) Decode(c Reader) ([]byte, MsgType, int, error) {

	minSize := MagicNumberStartLength + MsgTypeLength
	if c.InboundBuffered() < minSize {
		return nil, 0, 0, io.ErrShortBuffer
	}

	magicStart, err := c.Peek(MagicNumberStartLength)
	if err != nil {
		return nil, 0, 0, err
	}

	if !bytes.Equal(magicStart, MagicNumberStart) {
		d.Error("decode: invalid magic number start", zap.ByteString("act", magicStart), zap.ByteString("expect", MagicNumberStart), zap.Int("totalLen", c.InboundBuffered()))
		return nil, 0, 0, fmt.Errorf("invalid magic number start")
	}

	msgByteBuff, err := c.Peek(MagicNumberStartLength + MsgTypeLength)
	if err != nil {
		return nil, 0, 0, err
	}
	msgType := uint8(msgByteBuff[MagicNumberStartLength])

	if msgType == MsgTypeHeartbeat.Uint8() {
		_, _ = c.Discard(MagicNumberStartLength + MsgTypeLength)
		return []byte{MsgTypeHeartbeat.Uint8()}, MsgTypeHeartbeat, MagicNumberStartLength + MsgTypeLength, nil
	}

	contentLenBytes, err := c.Peek(MagicNumberStartLength + MsgTypeLength + MsgContentLength)
	if err != nil {
		return nil, 0, 0, err
	}

	contentLen := binary.BigEndian.Uint32(contentLenBytes[MagicNumberStartLength+MsgTypeLength:])

	totalSize := MagicNumberStartLength + MsgTypeLength + MsgContentLength + int(contentLen)
	buf, err := c.Peek(totalSize)
	if err != nil {
		return nil, 0, 0, err
	}

	contentBytes := make([]byte, contentLen)
	copy(contentBytes, buf[MagicNumberStartLength+MsgTypeLength+MsgContentLength:totalSize])

	msgLen := totalSize

	_, err = c.Discard(msgLen)
	if err != nil {
		d.Warn("discard error", zap.Error(err))
	}

	return contentBytes, MsgType(msgType), msgLen, nil
}

func (d *DefaultProto) Encode(data []byte, msgType MsgType) ([]byte, error) {
	if msgType == MsgTypeHeartbeat {
		msgData := make([]byte, MagicNumberStartLength+MsgTypeLength)
		copy(msgData, MagicNumberStart)
		msgData[MagicNumberStartLength] = msgType.Uint8()
		return msgData, nil
	}

	msgContentOffset := MsgTypeLength + MsgContentLength + len(MagicNumberStart)
	msgLen := msgContentOffset + len(data)

	msgData := make([]byte, msgLen)

	copy(msgData, MagicNumberStart)

	copy(msgData[MagicNumberStartLength:MagicNumberStartLength+1], []byte{msgType.Uint8()})

	binary.BigEndian.PutUint32(msgData[MagicNumberStartLength+MsgTypeLength:msgContentOffset], uint32(len(data)))

	copy(msgData[msgContentOffset:msgLen], data)

	return msgData, nil
}
