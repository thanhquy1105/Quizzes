package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"btaskee-quiz/quiz"
)

func main() {
	addr := "tcp://0.0.0.0:8080"
	server := quiz.NewQuizServer(addr)

	fmt.Printf("Starting Quiz Server on %s\n", addr)
	fmt.Printf("WebSocket Server started on :8081\n")
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
