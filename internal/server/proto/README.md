# 🚩 BTASKEE Binary Protocol

This document describes the custom binary framing protocol used for low-overhead, high-concurrency communication between the Btaskee Quiz Server and its clients.

[Bản tiếng Việt](README_VN.md)

## 🏗 Message Framing (Common Layer)

All messages follow a common header format. Multi-byte integers in the header are encoded in **Big Endian**.

| Offset | Length | Type   | Description |
|--------|--------|--------|-------------|
| 0      | 7      | string | **Magic Number**: Always `BTASKEE` |
| 7      | 1      | uint8  | **Message Type**: Defines the payload structure |
| 8      | 4      | uint32 | **Content Length**: Length of the following payload (Omitted for Heartbeats) |
| 12     | N      | bytes  | **Payload**: The message body |

---

## 🔢 Message Types

| Value | Name | Description |
|-------|------|-------------|
| 1     | `Connect` | Request to establish a session (with Auth token) |
| 2     | `Connack` | Server acknowledgement of a connection |
| 3     | `Request` | Generic client command (e.g., Join, Answer) |
| 4     | `Response`| Server response to a specific Request |
| 5     | `Heartbeat`| Keep-alive packet (8 bytes total: `BTASKEE` + `0x05`) |
| 6     | `Message` | Asynchronous server-to-client message (e.g., Leaderboard update) |
| 7     | `Batch`   | A bundle of multiple `Message` packets |

---

## 📦 Payload Structures

Most payload fields use **Little Endian** for consistency with Go's internal memory layout.

### 1. Connect (Type 1)
Used by clients to authenticate after a WebSocket handshake.
- `ID`: 8 bytes (uint64)
- `UsernameLen`: 2 bytes (uint16)
- `Username`: N bytes (string)
- `TokenLen`: 2 bytes (uint16)
- `Token`: N bytes (string)
- `BodyLen`: 4 bytes (uint32)
- `Body`: N bytes (optional data)

### 2. Connack (Type 2)
Sent by server to confirm authentication.
- `ID`: 8 bytes (uint64)
- `Status`: 1 byte (0: OK, 1: Error)
- `BodyLen`: 4 bytes (uint32)
- `Body`: N bytes (Error message or confirmation)

### 3. Request (Type 3)
Used for structured commands.
- `ID`: 8 bytes (uint64)
- `PathLen`: 2 bytes (uint16) - e.g., `join`, `answer`
- `Path`: N bytes (string)
- `BodyLen`: 4 bytes (uint32)
- `Body`: N bytes (JSON payload)

### 4. Response (Type 4)
- `ID`: 8 bytes (uint64) - Matches the Request ID.
- `Status`: 1 byte (0: OK, 1: Error, 2: NotFound, 3: AlreadyAnswered)
- `Timestamp`: 8 bytes (int64)
- `BodyLen`: 4 bytes (uint32)
- `Body`: N bytes (JSON result)

### 5. Message (Type 6)
- `ID`: 8 bytes (uint64)
- `MsgType`: 4 bytes (uint32)
- `Timestamp`: 8 bytes (uint64)
- `ContentLen`: 4 bytes (uint32)
- `Content`: N bytes (Payload)

---

## ⚡ Status Codes

| Value | Status |
|-------|--------|
| 0     | `OK` |
| 1     | `Error` |
| 2     | `NotFound` |
| 3     | `AlreadyAnswered` |