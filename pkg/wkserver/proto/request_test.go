package proto

import (
	"bytes"
	"fmt"
	"testing"
)

func TestRequest(t *testing.T) {

	var data []byte
	var err error
	t.Run("Marshal", func(t *testing.T) {

		req := &Request{
			Id:   12345,
			Path: "/test/path",
			Body: []byte("hello, world"),
		}

		data, err = req.Marshal()
		if err != nil {
			fmt.Println("Marshal error:", err)
			return
		}

	})

	t.Run("Unmarshal", func(t *testing.T) {

		var newReq Request
		if err := newReq.Unmarshal(data); err != nil {
			fmt.Println("Unmarshal error:", err)
			return
		}
	})
}

func TestConnect_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    Connect
		wantErr  bool
		expected Connect
	}{
		{
			name: "Basic test",
			input: Connect{
				Id:    1,
				Uid:   "user123",
				Token: "token123",
				Body:  []byte("hello world"),
			},
			wantErr: false,
			expected: Connect{
				Id:    1,
				Uid:   "user123",
				Token: "token123",
				Body:  []byte("hello world"),
			},
		},
		{
			name: "Empty Uid and Token",
			input: Connect{
				Id:    42,
				Uid:   "",
				Token: "",
				Body:  []byte("just body"),
			},
			wantErr: false,
			expected: Connect{
				Id:    42,
				Uid:   "",
				Token: "",
				Body:  []byte("just body"),
			},
		},
		{
			name: "Empty Body",
			input: Connect{
				Id:    1001,
				Uid:   "testuser",
				Token: "testtoken",
				Body:  []byte{},
			},
			wantErr: false,
			expected: Connect{
				Id:    1001,
				Uid:   "testuser",
				Token: "testtoken",
				Body:  []byte{},
			},
		},
		{
			name: "Large Body",
			input: Connect{
				Id:    9001,
				Uid:   "largeuser",
				Token: "largetoken",
				Body:  bytes.Repeat([]byte("a"), 1<<16),
			},
			wantErr: false,
			expected: Connect{
				Id:    9001,
				Uid:   "largeuser",
				Token: "largetoken",
				Body:  bytes.Repeat([]byte("a"), 1<<16),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			data, err := tt.input.Marshal()
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			var result Connect
			err = result.Unmarshal(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !compareConnect(result, tt.expected) {
				t.Errorf("Unmarshal() got = %+v, want = %+v", result, tt.expected)
			}
		})
	}
}

func compareConnect(a, b Connect) bool {
	return a.Id == b.Id && a.Uid == b.Uid && a.Token == b.Token && bytes.Equal(a.Body, b.Body)
}

func TestConnack_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    Connack
		wantErr  bool
		expected Connack
	}{
		{
			name: "Basic test",
			input: Connack{
				Id:     1,
				Status: 0,
				Body:   []byte("hello world"),
			},
			wantErr: false,
			expected: Connack{
				Id:     1,
				Status: 0,
				Body:   []byte("hello world"),
			},
		},
		{
			name: "Empty Body",
			input: Connack{
				Id:     42,
				Status: 1,
				Body:   []byte{},
			},
			wantErr: false,
			expected: Connack{
				Id:     42,
				Status: 1,
				Body:   []byte{},
			},
		},
		{
			name: "Large Body",
			input: Connack{
				Id:     9001,
				Status: 2,
				Body:   bytes.Repeat([]byte("a"), 1<<16),
			},
			wantErr: false,
			expected: Connack{
				Id:     9001,
				Status: 2,
				Body:   bytes.Repeat([]byte("a"), 1<<16),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			data, err := tt.input.Marshal()
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			var result Connack
			err = result.Unmarshal(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !compareConnack(result, tt.expected) {
				t.Errorf("Unmarshal() got = %+v, want = %+v", result, tt.expected)
			}
		})
	}
}

func compareConnack(a, b Connack) bool {
	return a.Id == b.Id && a.Status == b.Status && bytes.Equal(a.Body, b.Body)
}

func TestResponse_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    Response
		wantErr  bool
		expected Response
	}{
		{
			name: "Basic test",
			input: Response{
				Id:        1,
				Status:    StatusOK,
				Timestamp: 1637849938,
				Body:      []byte("hello world"),
			},
			wantErr: false,
			expected: Response{
				Id:        1,
				Status:    StatusOK,
				Timestamp: 1637849938,
				Body:      []byte("hello world"),
			},
		},
		{
			name: "Empty Body",
			input: Response{
				Id:        42,
				Status:    StatusNotFound,
				Timestamp: 1637849938,
				Body:      []byte{},
			},
			wantErr: false,
			expected: Response{
				Id:        42,
				Status:    StatusNotFound,
				Timestamp: 1637849938,
				Body:      []byte{},
			},
		},
		{
			name: "Large Body",
			input: Response{
				Id:        9001,
				Status:    StatusError,
				Timestamp: 1637849938,
				Body:      bytes.Repeat([]byte("a"), 1<<16),
			},
			wantErr: false,
			expected: Response{
				Id:        9001,
				Status:    StatusError,
				Timestamp: 1637849938,
				Body:      bytes.Repeat([]byte("a"), 1<<16),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			data, err := tt.input.Marshal()
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			var result Response
			err = result.Unmarshal(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !compareResponse(result, tt.expected) {
				t.Errorf("Unmarshal() got = %+v, want = %+v", result, tt.expected)
			}
		})
	}
}

func compareResponse(a, b Response) bool {
	return a.Id == b.Id && a.Status == b.Status && a.Timestamp == b.Timestamp && bytes.Equal(a.Body, b.Body)
}

func TestMessage_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    Message
		wantErr  bool
		expected Message
	}{
		{
			name: "Basic test",
			input: Message{
				Id:        1,
				MsgType:   1001,
				Content:   []byte("hello world"),
				Timestamp: 1637849938,
			},
			wantErr: false,
			expected: Message{
				Id:        1,
				MsgType:   1001,
				Content:   []byte("hello world"),
				Timestamp: 1637849938,
			},
		},
		{
			name: "Empty Content",
			input: Message{
				Id:        42,
				MsgType:   2001,
				Content:   []byte{},
				Timestamp: 1637849938,
			},
			wantErr: false,
			expected: Message{
				Id:        42,
				MsgType:   2001,
				Content:   []byte{},
				Timestamp: 1637849938,
			},
		},
		{
			name: "Large Content",
			input: Message{
				Id:        9001,
				MsgType:   3001,
				Content:   bytes.Repeat([]byte("a"), 1<<16),
				Timestamp: 1637849938,
			},
			wantErr: false,
			expected: Message{
				Id:        9001,
				MsgType:   3001,
				Content:   bytes.Repeat([]byte("a"), 1<<16),
				Timestamp: 1637849938,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			data, err := tt.input.Marshal()
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			var result Message
			err = result.Unmarshal(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !compareMessage(result, tt.expected) {
				t.Errorf("Unmarshal() got = %+v, want = %+v", result, tt.expected)
			}
		})
	}
}

func compareMessage(a, b Message) bool {
	return a.Id == b.Id && a.MsgType == b.MsgType && a.Timestamp == b.Timestamp && bytes.Equal(a.Content, b.Content)
}

func TestBatchMessage_EncodeDecodeBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    BatchMessage
		wantErr  bool
		expected BatchMessage
	}{
		{
			name: "Empty batch",
			input: BatchMessage{
				Messages: []*Message{},
				Count:    0,
			},
			wantErr: false,
			expected: BatchMessage{
				Messages: []*Message{},
				Count:    0,
			},
		},
		{
			name: "Single message",
			input: BatchMessage{
				Messages: []*Message{
					{
						Id:        1,
						MsgType:   uint32(MsgTypeRequest),
						Content:   []byte("test message"),
						Timestamp: 1637849938,
					},
				},
				Count: 1,
			},
			wantErr: false,
			expected: BatchMessage{
				Messages: []*Message{
					{
						Id:        1,
						MsgType:   uint32(MsgTypeRequest),
						Content:   []byte("test message"),
						Timestamp: 1637849938,
					},
				},
				Count: 1,
			},
		},
		{
			name: "Multiple messages",
			input: BatchMessage{
				Messages: []*Message{
					{
						Id:        1,
						MsgType:   uint32(MsgTypeRequest),
						Content:   []byte("message 1"),
						Timestamp: 1637849938,
					},
					{
						Id:        2,
						MsgType:   uint32(MsgTypeResp),
						Content:   []byte("message 2"),
						Timestamp: 1637849939,
					},
					{
						Id:        3,
						MsgType:   uint32(MsgTypeHeartbeat),
						Content:   []byte{},
						Timestamp: 1637849940,
					},
				},
				Count: 3,
			},
			wantErr: false,
			expected: BatchMessage{
				Messages: []*Message{
					{
						Id:        1,
						MsgType:   uint32(MsgTypeRequest),
						Content:   []byte("message 1"),
						Timestamp: 1637849938,
					},
					{
						Id:        2,
						MsgType:   uint32(MsgTypeResp),
						Content:   []byte("message 2"),
						Timestamp: 1637849939,
					},
					{
						Id:        3,
						MsgType:   uint32(MsgTypeHeartbeat),
						Content:   []byte{},
						Timestamp: 1637849940,
					},
				},
				Count: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			data, err := tt.input.Encode()
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			var result BatchMessage
			err = result.Decode(data)
			if err != nil {
				t.Errorf("Decode() error = %v", err)
				return
			}

			if !compareBatchMessage(result, tt.expected) {
				t.Errorf("Decode() got = %+v, want = %+v", result, tt.expected)
			}
		})
	}
}

func compareBatchMessage(a, b BatchMessage) bool {
	if a.Count != b.Count {
		return false
	}

	if len(a.Messages) != len(b.Messages) {
		return false
	}

	for i, msgA := range a.Messages {
		msgB := b.Messages[i]
		if !compareMessage(*msgA, *msgB) {
			return false
		}
	}

	return true
}

func TestBatchMessage_EdgeCases(t *testing.T) {
	t.Run("Large batch", func(t *testing.T) {

		messageCount := 1000
		messages := make([]*Message, messageCount)
		for i := 0; i < messageCount; i++ {
			messages[i] = &Message{
				Id:        uint64(i),
				MsgType:   uint32(MsgTypeRequest),
				Content:   []byte(fmt.Sprintf("message %d", i)),
				Timestamp: uint64(1637849938 + i),
			}
		}

		batchMsg := BatchMessage{
			Messages: messages,
			Count:    uint32(messageCount),
		}

		data, err := batchMsg.Encode()
		if err != nil {
			t.Errorf("Encode() error = %v", err)
			return
		}

		var result BatchMessage
		err = result.Decode(data)
		if err != nil {
			t.Errorf("Decode() error = %v", err)
			return
		}

		if result.Count != uint32(messageCount) {
			t.Errorf("Count mismatch: got %d, want %d", result.Count, messageCount)
		}

		if len(result.Messages) != messageCount {
			t.Errorf("Messages length mismatch: got %d, want %d", len(result.Messages), messageCount)
		}

		for i := 0; i < 5; i++ {
			if !compareMessage(*result.Messages[i], *messages[i]) {
				t.Errorf("Message %d mismatch", i)
			}
		}

		for i := messageCount - 5; i < messageCount; i++ {
			if !compareMessage(*result.Messages[i], *messages[i]) {
				t.Errorf("Message %d mismatch", i)
			}
		}
	})

	t.Run("Messages with large content", func(t *testing.T) {

		largeContent := bytes.Repeat([]byte("A"), 64*1024)
		messages := []*Message{
			{
				Id:        1,
				MsgType:   uint32(MsgTypeRequest),
				Content:   largeContent,
				Timestamp: 1637849938,
			},
			{
				Id:        2,
				MsgType:   uint32(MsgTypeResp),
				Content:   bytes.Repeat([]byte("B"), 32*1024),
				Timestamp: 1637849939,
			},
		}

		batchMsg := BatchMessage{
			Messages: messages,
			Count:    2,
		}

		data, err := batchMsg.Encode()
		if err != nil {
			t.Errorf("Encode() error = %v", err)
			return
		}

		var result BatchMessage
		err = result.Decode(data)
		if err != nil {
			t.Errorf("Decode() error = %v", err)
			return
		}

		if !compareBatchMessage(result, batchMsg) {
			t.Errorf("Large content batch message mismatch")
		}
	})

	t.Run("Mixed message types", func(t *testing.T) {

		messages := []*Message{
			{
				Id:        1,
				MsgType:   uint32(MsgTypeConnect),
				Content:   []byte("connect data"),
				Timestamp: 1637849938,
			},
			{
				Id:        2,
				MsgType:   uint32(MsgTypeConnack),
				Content:   []byte("connack data"),
				Timestamp: 1637849939,
			},
			{
				Id:        3,
				MsgType:   uint32(MsgTypeRequest),
				Content:   []byte("request data"),
				Timestamp: 1637849940,
			},
			{
				Id:        4,
				MsgType:   uint32(MsgTypeResp),
				Content:   []byte("response data"),
				Timestamp: 1637849941,
			},
			{
				Id:        5,
				MsgType:   uint32(MsgTypeHeartbeat),
				Content:   []byte{},
				Timestamp: 1637849942,
			},
			{
				Id:        6,
				MsgType:   uint32(MsgTypeMessage),
				Content:   []byte("message data"),
				Timestamp: 1637849943,
			},
		}

		batchMsg := BatchMessage{
			Messages: messages,
			Count:    uint32(len(messages)),
		}

		data, err := batchMsg.Encode()
		if err != nil {
			t.Errorf("Encode() error = %v", err)
			return
		}

		var result BatchMessage
		err = result.Decode(data)
		if err != nil {
			t.Errorf("Decode() error = %v", err)
			return
		}

		if !compareBatchMessage(result, batchMsg) {
			t.Errorf("Mixed message types batch mismatch")
		}
	})
}

func TestBatchMessage_ErrorHandling(t *testing.T) {
	t.Run("Decode empty data", func(t *testing.T) {
		var batchMsg BatchMessage
		err := batchMsg.Decode([]byte{})
		if err == nil {
			t.Error("Expected error for empty data, got nil")
		}
		if err.Error() != "batch message data too short" {
			t.Errorf("Expected 'batch message data too short', got '%s'", err.Error())
		}
	})

	t.Run("Decode insufficient data", func(t *testing.T) {
		var batchMsg BatchMessage

		err := batchMsg.Decode([]byte{0x01, 0x02})
		if err == nil {
			t.Error("Expected error for insufficient data, got nil")
		}
		if err.Error() != "batch message data too short" {
			t.Errorf("Expected 'batch message data too short', got '%s'", err.Error())
		}
	})

	t.Run("Decode truncated message data", func(t *testing.T) {

		originalMsg := &Message{
			Id:        1,
			MsgType:   uint32(MsgTypeRequest),
			Content:   []byte("test message"),
			Timestamp: 1637849938,
		}

		batchMsg := BatchMessage{
			Messages: []*Message{originalMsg},
			Count:    1,
		}

		data, err := batchMsg.Encode()
		if err != nil {
			t.Fatalf("Encode() error = %v", err)
		}

		if len(data) > 10 {
			truncatedData := data[:len(data)-10]

			var result BatchMessage
			err = result.Decode(truncatedData)
			if err == nil {
				t.Error("Expected error for truncated data, got nil")
			}
		}
	})

	t.Run("Count mismatch", func(t *testing.T) {

		originalMsg := &Message{
			Id:        1,
			MsgType:   uint32(MsgTypeRequest),
			Content:   []byte("test"),
			Timestamp: 1637849938,
		}

		msgData, err := originalMsg.Encode()
		if err != nil {
			t.Fatalf("Message encode error = %v", err)
		}

		data := make([]byte, 4+len(msgData))

		data[0] = 2
		data[1] = 0
		data[2] = 0
		data[3] = 0

		copy(data[4:], msgData)

		var result BatchMessage
		err = result.Decode(data)
		if err == nil {
			t.Error("Expected error for count mismatch, got nil")
		}
	})

	t.Run("Invalid message in batch", func(t *testing.T) {

		data := []byte{

			1, 0, 0, 0,

			1, 2, 3,
		}

		var result BatchMessage
		err := result.Decode(data)
		if err == nil {
			t.Error("Expected error for invalid message, got nil")
		}
	})
}

func TestBatchMessage_Size(t *testing.T) {
	tests := []struct {
		name     string
		input    BatchMessage
		expected int
	}{
		{
			name: "Empty batch",
			input: BatchMessage{
				Messages: []*Message{},
				Count:    0,
			},
			expected: 4,
		},
		{
			name: "Single message",
			input: BatchMessage{
				Messages: []*Message{
					{
						Id:        1,
						MsgType:   1,
						Content:   []byte("test"),
						Timestamp: 1637849938,
					},
				},
				Count: 1,
			},
			expected: 4 + (8 + 4 + 8 + 4 + 4),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := tt.input.Size()
			if size != tt.expected {
				t.Errorf("Size() = %d, want %d", size, tt.expected)
			}

			data, err := tt.input.Encode()
			if err != nil {
				t.Errorf("Encode() error = %v", err)
				return
			}

			if len(data) != size {
				t.Errorf("Encoded data length %d != Size() %d", len(data), size)
			}
		})
	}
}

func TestBatchMessage_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	messageCount := 10000
	messages := make([]*Message, messageCount)
	for i := 0; i < messageCount; i++ {
		messages[i] = &Message{
			Id:        uint64(i),
			MsgType:   uint32(MsgTypeRequest),
			Content:   []byte("performance test message"),
			Timestamp: uint64(1637849938 + i),
		}
	}

	batchMsg := BatchMessage{
		Messages: messages,
		Count:    uint32(messageCount),
	}

	data, err := batchMsg.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	t.Logf("Encoded %d messages, total size: %d bytes", messageCount, len(data))

	var result BatchMessage
	err = result.Decode(data)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if result.Count != uint32(messageCount) {
		t.Errorf("Count mismatch: got %d, want %d", result.Count, messageCount)
	}

	if len(result.Messages) != messageCount {
		t.Errorf("Messages length mismatch: got %d, want %d", len(result.Messages), messageCount)
	}
}

func BenchmarkBatchMessage_Encode(b *testing.B) {

	sizes := []int{1, 10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {

			messages := make([]*Message, size)
			for i := 0; i < size; i++ {
				messages[i] = &Message{
					Id:        uint64(i),
					MsgType:   uint32(MsgTypeRequest),
					Content:   []byte("benchmark test message"),
					Timestamp: uint64(1637849938 + i),
				}
			}

			batchMsg := BatchMessage{
				Messages: messages,
				Count:    uint32(size),
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := batchMsg.Encode()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkBatchMessage_Decode(b *testing.B) {

	sizes := []int{1, 10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {

			messages := make([]*Message, size)
			for i := 0; i < size; i++ {
				messages[i] = &Message{
					Id:        uint64(i),
					MsgType:   uint32(MsgTypeRequest),
					Content:   []byte("benchmark test message"),
					Timestamp: uint64(1637849938 + i),
				}
			}

			batchMsg := BatchMessage{
				Messages: messages,
				Count:    uint32(size),
			}

			data, err := batchMsg.Encode()
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result BatchMessage
				err := result.Decode(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkBatchMessage_EncodeDecodeRoundtrip(b *testing.B) {

	messages := make([]*Message, 50)
	for i := 0; i < 50; i++ {
		messages[i] = &Message{
			Id:        uint64(i),
			MsgType:   uint32(MsgTypeRequest),
			Content:   []byte("roundtrip test message"),
			Timestamp: uint64(1637849938 + i),
		}
	}

	batchMsg := BatchMessage{
		Messages: messages,
		Count:    50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		data, err := batchMsg.Encode()
		if err != nil {
			b.Fatal(err)
		}

		var result BatchMessage
		err = result.Decode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
