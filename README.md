# Real-Time Quiz Application Demo

This project demonstrates a real-time quiz feature using a Go-based server (`wkserver`) and a ReactJS frontend connected via WebSockets using a custom binary protocol.

## Features
- **Real-time Scoring**: Points are updated instantly across all participants.
- **Live Leaderboard**: Standings update automatically when scores change.
- **WebSocket Protocol**: Custom binary protocol (`WUKONG`) for efficient communication.

## How to Run

### 1. Start the Go Server
Navigate to the root and run the quiz server:
```bash
go run cmd/quizserver/main.go
```
The server will start on:
- TCP: `:8080` (Standard binary protocol)
- WebSocket: `:8080` (WebSocket binary protocol at `/ws` - default)

### 2. Run the React Frontend
Navigate to `demo-web` and start the development server:
```bash
cd demo-web
npm install
npm run dev
```
Open [http://localhost:5173](http://localhost:5173) in your browser.

## Technical Details
- **Backend**: Built with `gnet` (TCP) and `wknet` (WS).
- **Frontend**: Built with Vite, React, TypeScript, and Tailwind CSS.
- **Protocol**: Binary framing with Magic Number `WUKONG`.
