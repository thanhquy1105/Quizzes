package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"btaskee-quiz/internal/model"
	"btaskee-quiz/internal/repository"
	"btaskee-quiz/pkg/token"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	userStore  repository.UserStore
	quizStore  repository.QuizStore
	tokenStore repository.TokenStore
	tokenMaker token.IMaker
	accessDur  time.Duration
	refreshDur time.Duration
}

func NewHandler(userStore repository.UserStore, quizStore repository.QuizStore, tokenStore repository.TokenStore, tokenMaker token.IMaker, accessDur time.Duration, refreshDur time.Duration) *Handler {
	return &Handler{
		userStore:  userStore,
		quizStore:  quizStore,
		tokenStore: tokenStore,
		tokenMaker: tokenMaker,
		accessDur:  accessDur,
		refreshDur: refreshDur,
	}
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

type LoginResp struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userStore.GetByUsername(c.Request.Context(), req.Username)
	if err != nil {
		// Create new user
		user = &model.User{
			Username: req.Username,
			Name:     req.Name,
		}
		if err := h.userStore.Save(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if user.Name != req.Name {
		// Update name
		user.Name = req.Name
		if err := h.userStore.Save(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	}

	accessToken, accessPayload, err := h.tokenMaker.CreateToken(user.Username, h.accessDur, token.TokenTypeSessionToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create access token"})
		return
	}

	refreshToken, refreshPayload, err := h.tokenMaker.CreateToken(user.Username, h.refreshDur, token.TokenTypeRefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refresh token"})
		return
	}

	err = h.tokenStore.Save(c.Request.Context(), accessToken, user.Username, h.accessDur, token.TokenTypeSessionToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save access token session"})
		return
	}

	err = h.tokenStore.Save(c.Request.Context(), refreshToken, user.Username, h.refreshDur, token.TokenTypeRefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save refresh token session"})
		return
	}

	c.JSON(http.StatusOK, LoginResp{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
	})
}

type ListQuizzesResp struct {
	Quizzes []model.Quiz `json:"quizzes"`
}

func (h *Handler) ListQuizzes(c *gin.Context) {
	quizzes, err := h.quizStore.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListQuizzesResp{
		Quizzes: quizzes,
	})
}

func (h *Handler) GetDetailedQuiz(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quiz id"})
		return
	}

	quiz, err := h.quizStore.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quiz not found"})
		return
	}

	c.JSON(http.StatusOK, quiz)
}

type ListSessionsResp struct {
	Sessions []model.QuizSession `json:"sessions"`
}

func (h *Handler) ListSessions(c *gin.Context) {
	sessions, err := h.quizStore.ListSessions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListSessionsResp{
		Sessions: sessions,
	})
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload, err := h.tokenMaker.VerifyToken(req.RefreshToken, token.TokenTypeRefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	exists, err := h.tokenStore.Exists(c.Request.Context(), req.RefreshToken, token.TokenTypeRefreshToken)
	if err != nil || !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token is revoked or expired"})
		return
	}

	_ = h.tokenStore.Delete(c.Request.Context(), req.RefreshToken, token.TokenTypeRefreshToken)

	accessToken, accessPayload, err := h.tokenMaker.CreateToken(payload.Username, h.accessDur, token.TokenTypeSessionToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create access token"})
		return
	}

	refreshToken, refreshPayload, err := h.tokenMaker.CreateToken(payload.Username, h.refreshDur, token.TokenTypeRefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refresh token"})
		return
	}

	err = h.tokenStore.Save(c.Request.Context(), accessToken, payload.Username, h.accessDur, token.TokenTypeSessionToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save access token session"})
		return
	}

	err = h.tokenStore.Save(c.Request.Context(), refreshToken, payload.Username, h.refreshDur, token.TokenTypeRefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save refresh token session"})
		return
	}

	c.JSON(http.StatusOK, LoginResp{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
	})
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is not provided"})
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		accessToken := fields[1]
		payload, err := h.tokenMaker.VerifyToken(accessToken, token.TokenTypeSessionToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		exists, err := h.tokenStore.Exists(c.Request.Context(), accessToken, token.TokenTypeSessionToken)
		if err != nil || !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked or expired"})
			return
		}

		c.Set("user_id", payload.Username)
		c.Next()
	}
}
