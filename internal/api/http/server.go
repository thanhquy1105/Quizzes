package http

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
	addr   string
}

func NewServer(addr string, handler *Handler) *Server {
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.POST("/login", handler.Login)
	r.POST("/refresh", handler.Refresh)

	auth := r.Group("/").Use(handler.AuthMiddleware())
	auth.GET("/quizzes", handler.ListQuizzes)
	auth.GET("/sessions", handler.ListSessions)
	// auth.GET("/quizzes/:id", handler.GetDetailedQuiz)

	return &Server{
		engine: r,
		addr:   addr,
	}
}

func (s *Server) Start() error {
	return s.engine.Run(s.addr)
}
