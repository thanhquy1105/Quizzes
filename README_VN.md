# 🏆 Btaskee Real-Time Quiz Server

Hệ thống Quiz thời gian thực hiệu năng cao, mở rộng linh hoạt được thiết kế để xử lý hàng ngàn người tham gia cùng lúc với độ trễ phản hồi cực thấp. Dự án cung cấp nền tảng backend mạnh mẽ cho các cuộc thi quiz trực tiếp, nổi bật với khả năng xác thực atomic và cập nhật bảng xếp hạng tức thời.

[English Version](README.md)

---

## 🌟 Tổng quan dự án

**Btaskee Quiz Server** được xây dựng cho các phiên chơi game có lượng truy cập đồng thời lớn. Hệ thống tận dụng các công nghệ hiện đại để đảm bảo tính toàn vẹn dữ liệu trong khi vẫn duy trì hiệu suất cực cao. Hệ thống sử dụng mô hình **"Redis Checklist"** độc đáo để ngăn chặn việc nộp bài trùng lặp và giảm tải các thao tác đọc nặng nề sang Redis, đồng thời đảm bảo dữ liệu được lưu trữ vĩnh viễn trong MySQL.

### 🚀 Điểm nổi bật
- **10k+ Người dùng đồng thời**: Sử dụng `gnet` và tối ưu hóa WebSocket để xử lý quy mô lớn.
- **Xác thực dưới 1ms**: Redis Sets (O(1)) giúp kiểm tra câu trả lời ngay lập tức, giảm 100% tải DB cho luồng xử lý chính.
- **Lưu trữ hai tầng (Dual-Tier)**: Kết hợp Redis để đạt tốc độ tối đa và MySQL để đảm bảo tính toàn vẹn giao dịch.
- **Bảng xếp hạng toàn cầu thời gian thực**: Cập nhật thứ hạng tức thì qua Redis Sorted Sets (ZSet).
- **Cơ chế tự phục hồi**: Tự động hâm nóng cache (warming) khi người dùng tham gia lại, đảm bảo tính nhất quán ngay cả sau khi restart.
- **Giao thức Binary**: Sử dụng khung truyền tin (framing) `BTASKEE` tùy chỉnh trên WebSockets để giao tiếp hiệu quả, tối ưu băng thông. Xem [Tài liệu Protocol](internal/server/proto/README.md).
- **Sẵn sàng cho Docker**: Khởi chạy toàn bộ hệ sinh thái (Frontend + Backend + DB) chỉ với một lệnh duy nhất.

---

## 🛠 Công nghệ sử dụng

- **Backend**: Go (Golang)
- **Frontend**: Vite + React + TypeScript (nằm trong thư mục `demo-web`)
- **Networking**: `gnet` (TCP hiệu năng cao) & `Gorilla WebSocket`
- **Caching**: Redis 7 (Sorted Sets, Strings, Sets)
- **Database**: MySQL 8 (GORM)
- **Containerization**: Docker & Docker Compose

---

## 📦 Hướng dẫn chạy chương trình

### Cách 1: Sử dụng Docker (Nhanh nhất)
Tự động khởi chạy toàn bộ hệ thống (Database, Redis, Backend, và Frontend):

```bash
# Khởi chạy tất cả dịch vụ
make docker-up

# Xem log hệ thống
make docker-logs

# Dừng tất cả dịch vụ
make docker-down
```
- **Frontend**: `http://localhost:3000`
- **Backend API**: `http://localhost:8082`

---

### Cách 2: Thiết lập thủ công (Cho phát triển)

#### 1. Yêu cầu hệ thống
- Go 1.25+
- Node.js 20+
- Redis 6+ & MySQL 8+ đã được cài đặt và chạy cục bộ.

#### 2. Thiết lập cơ sở hạ tầng
Nếu bạn không có sẵn DB cục bộ, bạn có thể dùng Makefile để chỉ khởi chạy các database:
```bash
# Chỉ khởi chạy MySQL và Redis
docker compose up -d mysql redis
```

#### 3. Cài đặt Backend
```bash
go mod tidy
make run
```

#### 4. Cài đặt Frontend
```bash
cd demo-web
npm install
npm run dev
```

---

## 📂 Cơ cấu dự án

- `cmd/quizserver/`: Điểm khởi đầu chính của ứng dụng.
- `internal/quiz/`: Logic nghiệp vụ lõi (WebSocket, quản lý Session).
- `internal/api/http/`: Các endpoint REST cho quản trị và công khai.
- `internal/server/`: Framework mạng, giao thức binary và các chỉ số đo lường (metrics).
- `internal/repository/`: Lớp dữ liệu (Triển khai MySQL GORM & Redis cache).
- `demo-web/`: Frontend demo dựa trên React.
- `pkg/`: Các tiện ích có thể tái sử dụng (JWT, Errors, v.v.).

---

## 🛡 Công cụ quản trị

### 🔄 Nạp lại bảng xếp hạng (Reload Leaderboard)
Nếu cache bị mất hoặc cần đồng bộ lại dữ liệu từ MySQL (Yêu cầu `Authorization: Bearer <TOKEN>`):

```bash
# Nạp lại dữ liệu cho một phiên cụ thể (thay {CODE} bằng mã phiên quiz)
POST http://localhost:8082/sessions/{CODE}/reload-leaderboard

# Nạp lại dữ liệu cho tất cả các phiên đang hoạt động
POST http://localhost:8082/sessions/reload-leaderboard
```

---

## 📝 Giấy phép
Chỉ phục vụ mục đích demo nội bộ.
