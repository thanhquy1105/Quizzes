package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"btaskee-quiz/internal/api/http"
	"btaskee-quiz/internal/config"
	"btaskee-quiz/internal/quiz"
	mysqlrepo "btaskee-quiz/internal/repository/mysql"
	redisrepo "btaskee-quiz/internal/repository/redis"
	"btaskee-quiz/pkg/token"

	goredis "github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load("config.yml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// Initialize stores
	db, err := mysqlrepo.NewDB(&cfg.MySQL)
	if err != nil {
		fmt.Printf("Failed to connect to mysql: %v\n", err)
		return
	}
	dbUserStore := mysqlrepo.NewUserStore(db)
	dbQuizStore := mysqlrepo.NewQuizStore(db)

	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	tokenStore := redisrepo.NewTokenStore(rdb)

	// Wrap stores with Redis cache
	userStore := redisrepo.NewUserCache(rdb, dbUserStore)
	quizStoreCached := redisrepo.NewQuizCache(rdb, dbQuizStore)
	lbStore := redisrepo.NewLeaderboardStore(rdb, dbQuizStore)

	tokenMaker, err := token.NewJWTMaker(cfg.Token.SecretKey)
	if err != nil {
		fmt.Printf("Failed to create token maker: %v\n", err)
		return
	}

	// Start Gin HTTP Server
	httpHandler := http.NewHandler(userStore, quizStoreCached, lbStore, tokenStore, tokenMaker, cfg.Token.AccessTokenDuration, cfg.Token.RefreshTokenDuration)
	httpServer := http.NewServer(cfg.Server.HTTPAddr, httpHandler)
	go func() {
		fmt.Printf("HTTP server starting on %s\n", cfg.Server.HTTPAddr)
		if err := httpServer.Start(); err != nil {
			fmt.Printf("HTTP server failed: %v\n", err)
		}
	}()

	// Start WebSocket Server
	wsServer := quiz.NewQuizServer(cfg, rdb, tokenStore, dbQuizStore, dbUserStore, tokenMaker)
	go func() {
		fmt.Printf("WebSocket server starting on %s\n", cfg.Server.WSAddr)
		if err := wsServer.Start(); err != nil {
			fmt.Printf("WebSocket server failed: %v\n", err)
		}
	}()

	cmdCh := make(chan os.Signal, 1)
	signal.Notify(cmdCh, os.Interrupt, syscall.SIGTERM)
	<-cmdCh

	fmt.Println("Shutting down Quiz Server...")
	wsServer.Server.Stop()
}
