# 🏆 Btaskee Real-Time Quiz Server

A high-performance, scalable real-time quiz ecosystem designed to handle thousands of concurrent participants with sub-millisecond validation latency. This project provides a robust backend for competitive live quizzes, featuring atomic validation and instant leaderboard updates.

[Bản tiếng Việt](README_VN.md)

---

## 🌟 Project Overview

The **Btaskee Quiz Server** is built for high-concurrency gaming sessions. It leverages a modern tech stack to ensure data integrity while maintaining extreme performance. The system uses a unique **"Redis Checklist"** pattern to prevent double-submissions and offloads heavy read operations to Redis, while ensuring permanent persistence in MySQL.

### 🚀 Key Highlights
- **10k+ Concurrent Users**: Powered by `gnet` and optimized WebSocket handling for massive scale.
- **Sub-MS Validation**: Redis Sets (O(1)) provide instant answer checking, reducing DB load by 100% on the hot-path.
- **Dual-Tier Persistence**: Hybrid model using Redis for live speed and MySQL for transactional integrity.
- **Real-Time Global Rankings**: Instant updates via Redis Sorted Sets (ZSet).
- **Self-Healing Infrastructure**: Cache warming during joins ensures consistency even after restarts.
- **Binary Protocol**: Custom `BTASKEE` framing over WebSockets for efficient, low-overhead communication. See [Protocol Documentation](internal/server/proto/README.md).
- **Docker-Ready**: Launch the entire ecosystem (Frontend + Backend + DB) with a single command.

---

## 🛠 Tech Stack

- **Backend**: Go (Golang)
- **Frontend**: Vite + React + TypeScript (located in `demo-web`)
- **Networking**: `gnet` (High-performance TCP) & `Gorilla WebSocket`
- **Caching**: Redis 7 (Sorted Sets, Strings, Sets)
- **Database**: MySQL 8 (GORM)
- **Containerization**: Docker & Docker Compose

---

## 📦 Getting Started

### Method 1: Docker (Quickest)
Launch the entire stack (Database, Redis, Backend, and Frontend) automatically:

```bash
# Start all services
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```
- **Frontend**: `http://localhost:3000`
- **Backend API**: `http://localhost:8082`

---

### Method 2: Manual Setup (Development)

#### 1. Prerequisites
- Go 1.25+
- Node.js 20+
- Redis 6+ & MySQL 8+ running locally

#### 2. Infrastructure Setup
If you don't have DBs installed locally, you can use the Makefile to start only the infrastructure:
```bash
# This starts only MySQL and Redis
docker compose up -d mysql redis
```

#### 3. Backend Setup
```bash
go mod tidy
make run
```

#### 4. Frontend Setup
```bash
cd demo-web
npm install
npm run dev
```

---

## 📂 Project Structure

- `cmd/quizserver/`: Main application entry point.
- `internal/quiz/`: Core domain logic (WebSocket handlers, Session management).
- `internal/api/http/`: Administrative and public REST endpoints.
- `internal/server/`: Networking framework, binary protocol, and metrics.
- `internal/repository/`: Data layer (MySQL GORM & Redis cache implementations).
- `demo-web/`: React-based demonstration frontend.
- `pkg/`: Reusable utilities (JWT, Errors, etc.).

---

## 🛡 Administrative Tools

### 🔄 Reload Leaderboard
If the cache is lost or data needs re-syncing from MySQL (Requires `Authorization: Bearer <TOKEN>`):

```bash
# Reload a specific session (replace {CODE} with the session code)
POST http://localhost:8082/sessions/{CODE}/reload-leaderboard

# Reload all active sessions
POST http://localhost:8082/sessions/reload-leaderboard
```

---

## 📝 License
Internal demo purposes only.
