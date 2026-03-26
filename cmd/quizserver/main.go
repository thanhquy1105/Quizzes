package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"btaskee-quiz/internal/config"
	"btaskee-quiz/internal/quiz"
)

func main() {
	cfg, err := config.Load("config.yml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	server := quiz.NewQuizServer(cfg)

	fmt.Printf("TCP server starting on %s\n", cfg.Server.TCPAddr)
	fmt.Printf("WebSocket server starting on %s\n", cfg.Server.GorillaWSAddr)
	if err := server.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}

	cmdCh := make(chan os.Signal, 1)
	signal.Notify(cmdCh, os.Interrupt, syscall.SIGTERM)
	<-cmdCh

	fmt.Println("Shutting down Quiz Server...")
	server.Server.Stop()
}
