package proto

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	IdSize          = 8
	UidLenSize      = 2
	TokenLenSize    = 2
	BodyLenSize     = 4
	ConnectMinSize  = IdSize + UidLenSize + TokenLenSize + BodyLenSize
	StatusSize      = 1
	PathLenSize     = 2
	RequestMinSize  = IdSize + PathLenSize + BodyLenSize
	ConnackMinSize  = IdSize + StatusSize + BodyLenSize
	TimestampSize   = 8
	ResponseMinSize = IdSize + StatusSize + TimestampSize + BodyLenSize
	MsgTypeSize     = 4
	ContentLenSize  = 4
	MessageMinSize  = IdSize + MsgTypeSize + TimestampSize + ContentLenSize
)

type Status uint8

const (
	StatusOK       Status = 0
	StatusError    Status = 1
	StatusNotFound Status = 2
)

type Request struct {
	Id   uint64
	Path string
	Body []byte
}

func (r *Request) Reset() {
	r.Id = 0
	r.Path = ""
	r.Body = r.Body[:0]
}

func (r *Request) Marshal() ([]byte, error) {
	pathLen := len(r.Path)
	bodyLen := len(r.Body)
	totalSize := IdSize + PathLenSize + pathLen + BodyLenSize + bodyLen

	buffer := make([]byte, totalSize)
	offset := 0

	binary.LittleEndian.PutUint64(buffer[offset:], r.Id)
	offset += IdSize

	binary.LittleEndian.PutUint16(buffer[offset:], uint16(pathLen))
	offset += PathLenSize
	copy(buffer[offset:], r.Path)
	offset += pathLen

	binary.LittleEndian.PutUint32(buffer[offset:], uint32(bodyLen))
	offset += BodyLenSize
	copy(buffer[offset:], r.Body)

	return buffer, nil
}

func (r *Request) Unmarshal(data []byte) error {
	if len(data) < RequestMinSize {
		return fmt.Errorf("data too short")
	}

	offset := 0
	r.Id = binary.LittleEndian.Uint64(data[offset:])
	offset += IdSize

	pathLen := binary.LittleEndian.Uint16(data[offset:])
	offset += PathLenSize
	if len(data) < offset+int(pathLen) {
		return fmt.Errorf("invalid path length")
	}
	r.Path = string(data[offset : offset+int(pathLen)])
	offset += int(pathLen)

	bodyLen := binary.LittleEndian.Uint32(data[offset:])
	offset += BodyLenSize
	if len(data) < offset+int(bodyLen) {
		return fmt.Errorf("invalid body length")
	}
	r.Body = data[offset : offset+int(bodyLen)]

	return nil
}

type Response struct {
	Id        uint64
	Status    Status
	Timestamp int64
	Body      []byte
}

func (r *Response) Marshal() ([]byte, error) {
	bodyLen := len(r.Body)

	totalSize := IdSize + StatusSize + TimestampSize + BodyLenSize + bodyLen
	buffer := make([]byte, totalSize)

	offset := 0

	binary.LittleEndian.PutUint64(buffer[offset:], r.Id)
	offset += IdSize

	buffer[offset] = uint8(r.Status)
	offset += StatusSize

	binary.LittleEndian.PutUint64(buffer[offset:], uint64(r.Timestamp))
	offset += TimestampSize

	binary.LittleEndian.PutUint32(buffer[offset:], uint32(bodyLen))
	offset += BodyLenSize
	copy(buffer[offset:], r.Body)

	return buffer, nil
}

func (r *Response) Unmarshal(data []byte) error {
	if len(data) < ResponseMinSize {
		return errors.New("data too short to decode")
	}

	offset := 0

	r.Id = binary.LittleEndian.Uint64(data[offset:])
	offset += IdSize

	r.Status = Status(data[offset])
	offset += StatusSize

	r.Timestamp = int64(binary.LittleEndian.Uint64(data[offset:]))
	offset += TimestampSize

	bodyLen := binary.LittleEndian.Uint32(data[offset:])
	offset += BodyLenSize
	if len(data) < offset+int(bodyLen) {
		return errors.New("invalid Body length")
	}
	r.Body = data[offset : offset+int(bodyLen)]

	return nil
}

type Connect struct {
	Id    uint64
	Uid   string
	Token string
	Body  []byte
}

func (c *Connect) Marshal() ([]byte, error) {
	uidLen := len(c.Uid)
	tokenLen := len(c.Token)
	bodyLen := len(c.Body)

	totalSize := IdSize + UidLenSize + uidLen + TokenLenSize + tokenLen + BodyLenSize + bodyLen
	buffer := make([]byte, totalSize)

	offset := 0

	binary.LittleEndian.PutUint64(buffer[offset:], c.Id)
	offset += IdSize

	binary.LittleEndian.PutUint16(buffer[offset:], uint16(uidLen))
	offset += UidLenSize
	copy(buffer[offset:], c.Uid)
	offset += uidLen

	binary.LittleEndian.PutUint16(buffer[offset:], uint16(tokenLen))
	offset += TokenLenSize
	copy(buffer[offset:], c.Token)
	offset += tokenLen

	binary.LittleEndian.PutUint32(buffer[offset:], uint32(bodyLen))
	offset += BodyLenSize
	copy(buffer[offset:], c.Body)

	return buffer, nil
}

func (c *Connect) Unmarshal(data []byte) error {
	if len(data) < ConnectMinSize {
		return errors.New("data too short to decode")
	}

	offset := 0

	c.Id = binary.LittleEndian.Uint64(data[offset:])
	offset += IdSize

	uidLen := binary.LittleEndian.Uint16(data[offset:])
	offset += UidLenSize
	if len(data) < offset+int(uidLen) {
		return errors.New("invalid Uid length")
	}
	c.Uid = string(data[offset : offset+int(uidLen)])
	offset += int(uidLen)

	tokenLen := binary.LittleEndian.Uint16(data[offset:])
	offset += TokenLenSize
	if len(data) < offset+int(tokenLen) {
		return errors.New("invalid Token length")
	}
	c.Token = string(data[offset : offset+int(tokenLen)])
	offset += int(tokenLen)

	bodyLen := binary.LittleEndian.Uint32(data[offset:])
	offset += BodyLenSize
	if len(data) < offset+int(bodyLen) {
		return errors.New("invalid Body length")
	}
	c.Body = data[offset : offset+int(bodyLen)]

	return nil
}

type Connack struct {
	Id     uint64
	Status Status
	Body   []byte
}

func (c *Connack) Marshal() ([]byte, error) {
	bodyLen := len(c.Body)

	totalSize := IdSize + StatusSize + BodyLenSize + bodyLen
	buffer := make([]byte, totalSize)

	offset := 0

	binary.LittleEndian.PutUint64(buffer[offset:], c.Id)
	offset += IdSize

	buffer[offset] = uint8(c.Status)
	offset += StatusSize

	binary.LittleEndian.PutUint32(buffer[offset:], uint32(bodyLen))
	offset += BodyLenSize
	copy(buffer[offset:], c.Body)

	return buffer, nil
}

func (c *Connack) Unmarshal(data []byte) error {
	if len(data) < ConnackMinSize {
		return errors.New("data too short to decode")
	}

	offset := 0

	c.Id = binary.LittleEndian.Uint64(data[offset:])
	offset += IdSize

	c.Status = Status(data[offset])
	offset += StatusSize

	bodyLen := binary.LittleEndian.Uint32(data[offset:])
	offset += BodyLenSize
	if len(data) < offset+int(bodyLen) {
		return errors.New("invalid Body length")
	}
	c.Body = data[offset : offset+int(bodyLen)]

	return nil
}

type Message struct {
	Id        uint64
	MsgType   uint32
	Content   []byte
	Timestamp uint64
}

type BatchMessage struct {
	Messages []*Message
	Count    uint32
}

func (bm *BatchMessage) Size() int {
	size := 4
	for _, msg := range bm.Messages {
		size += msg.Size()
	}
	return size
}

func (bm *BatchMessage) Encode() ([]byte, error) {
	totalSize := bm.Size()
	data := make([]byte, totalSize)
	offset := 0

	binary.LittleEndian.PutUint32(data[offset:], bm.Count)
	offset += 4

	for _, msg := range bm.Messages {
		msgData, err := msg.Encode()
		if err != nil {
			return nil, err
		}
		copy(data[offset:], msgData)
		offset += len(msgData)
	}

	return data, nil
}

func (bm *BatchMessage) Decode(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("batch message data too short")
	}

	offset := 0

	bm.Count = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	bm.Messages = make([]*Message, 0, bm.Count)
	for i := uint32(0); i < bm.Count; i++ {
		if offset >= len(data) {
			return fmt.Errorf("batch message data truncated")
		}

		msg := &Message{}
		msgLen, err := msg.DecodeWithLength(data[offset:])
		if err != nil {
			return err
		}

		bm.Messages = append(bm.Messages, msg)
		offset += msgLen
	}

	return nil
}

func (m *Message) Size() int {

	contentLen := len(m.Content)
	totalSize := IdSize + MsgTypeSize + ContentLenSize + contentLen + TimestampSize

	return totalSize
}

func (m *Message) Marshal() ([]byte, error) {
	contentLen := len(m.Content)

	totalSize := IdSize + MsgTypeSize + TimestampSize + ContentLenSize + contentLen
	buffer := make([]byte, totalSize)

	offset := 0

	binary.LittleEndian.PutUint64(buffer[offset:], m.Id)
	offset += IdSize

	binary.LittleEndian.PutUint32(buffer[offset:], m.MsgType)
	offset += MsgTypeSize

	binary.LittleEndian.PutUint64(buffer[offset:], m.Timestamp)
	offset += TimestampSize

	binary.LittleEndian.PutUint32(buffer[offset:], uint32(contentLen))
	offset += ContentLenSize
	copy(buffer[offset:], m.Content)

	return buffer, nil
}

func (m *Message) Unmarshal(data []byte) error {
	if len(data) < MessageMinSize {
		return errors.New("data too short to decode")
	}

	offset := 0

	m.Id = binary.LittleEndian.Uint64(data[offset:])
	offset += IdSize

	m.MsgType = binary.LittleEndian.Uint32(data[offset:])
	offset += MsgTypeSize

	m.Timestamp = binary.LittleEndian.Uint64(data[offset:])
	offset += TimestampSize

	contentLen := binary.LittleEndian.Uint32(data[offset:])
	offset += ContentLenSize
	if len(data) < offset+int(contentLen) {
		return errors.New("invalid Content length")
	}
	m.Content = data[offset : offset+int(contentLen)]

	return nil
}

func (m *Message) Encode() ([]byte, error) {
	return m.Marshal()
}

func (m *Message) DecodeWithLength(data []byte) (int, error) {
	if len(data) < MessageMinSize {
		return 0, errors.New("data too short to decode")
	}

	offset := IdSize + MsgTypeSize + TimestampSize
	if len(data) < offset+ContentLenSize {
		return 0, errors.New("data too short to read content length")
	}

	contentLen := binary.LittleEndian.Uint32(data[offset:])
	totalLen := MessageMinSize + int(contentLen)

	if len(data) < totalLen {
		return 0, errors.New("data too short for complete message")
	}

	err := m.Unmarshal(data[:totalLen])
	if err != nil {
		return 0, err
	}

	return totalLen, nil
}
