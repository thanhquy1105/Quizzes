# 🚩 Giao thức Binary BTASKEE

Tài liệu này mô tả chi tiết về giao thức phân khung (framing) binary tùy chỉnh, được sử dụng để giao tiếp giữa Btaskee Quiz Server và các Client với hiệu năng cao và độ trễ thấp.

[English Version](README.md)

---

## 🏗 Phân khung tin nhắn (Common Layer)

Tất cả các tin nhắn đều tuân theo một định dạng Header chung. Các số nguyên nhiều byte trong Header được mã hóa theo kiểu **Big Endian**.

| Offset | Độ dài | Kiểu dữ liệu | Mô tả |
|--------|--------|--------------|-------|
| 0      | 7      | string       | **Magic Number**: Luôn là `BTASKEE` |
| 7      | 1      | uint8        | **MsgType**: Xác định cấu trúc của Payload bên dưới |
| 8      | 4      | uint32       | **Content Length**: Độ dài của Payload theo sau (Không có với Heartbeat) |
| 12     | N      | bytes        | **Payload**: Nội dung chính của tin nhắn |

---

## 🔢 Các loại tin nhắn (Message Types)

| Giá trị | Tên | Mô tả |
|---------|-----|-------|
| 1       | `Connect` | Yêu cầu thiết lập phiên (kèm Auth token) |
| 2       | `Connack` | Server xác nhận kết nối thành công |
| 3       | `Request` | Lệnh từ Client (ví dụ: Join, Answer) |
| 4       | `Response`| Phản hồi từ Server cho một Request cụ thể |
| 5       | `Heartbeat`| Gói tin duy trì kết nối (Tổng 8 byte: `BTASKEE` + `0x05`) |
| 6       | `Message` | Tin nhắn bất đồng bộ từ Server (ví dụ: Cập nhật BXH) |
| 7       | `Batch`   | Một gói gộp nhiều tin nhắn `Message` |

---

## 📦 Cấu trúc Payload

Hầu hết các trường trong Payload sử dụng **Little Endian** để tương thích tốt nhất với cách sắp xếp bộ nhớ của Go.

### 1. Connect (Loại 1)
Được Client sử dụng để xác thực sau khi bắt tay WebSocket thành công.
- `ID`: 8 bytes (uint64)
- `UsernameLen`: 2 bytes (uint16)
- `Username`: N bytes (string)
- `TokenLen`: 2 bytes (uint16)
- `Token`: N bytes (string)
- `BodyLen`: 4 bytes (uint32)
- `Body`: N bytes (dữ liệu tùy chọn)

### 2. Connack (Loại 2)
Server gửi để xác nhận việc xác thực thành công hay thất bại.
- `ID`: 8 bytes (uint64)
- `Status`: 1 byte (0: OK, 1: Error)
- `BodyLen`: 4 bytes (uint32)
- `Body`: N bytes (Thông báo lỗi hoặc xác nhận)

### 3. Request (Loại 3)
Dùng cho các lệnh có cấu trúc.
- `ID`: 8 bytes (uint64)
- `PathLen`: 2 bytes (uint16) - ví dụ: `join`, `answer`
- `Path`: N bytes (string)
- `BodyLen`: 4 bytes (uint32)
- `Body`: N bytes (Nội dung JSON)

### 4. Response (Loại 4)
- `ID`: 8 bytes (uint64) - Khớp với ID của Request tương ứng.
- `Status`: 1 byte (0: OK, 1: Error, 2: NotFound, 3: AlreadyAnswered)
- `Timestamp`: 8 bytes (int64)
- `BodyLen`: 4 bytes (uint32)
- `Body`: N bytes (Kết quả JSON)

### 5. Message (Loại 6)
- `ID`: 8 bytes (uint64)
- `MsgType`: 4 bytes (uint32)
- `Timestamp`: 8 bytes (uint64)
- `ContentLen`: 4 bytes (uint32)
- `Content`: N bytes (Payload)

---

## ⚡ Mã trạng thái (Status Codes)

| Giá trị | Trạng thái |
|---------|------------|
| 0       | `OK` |
| 1       | `Error` |
| 2       | `NotFound` |
| 3       | `AlreadyAnswered` |
