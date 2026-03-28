package quiz

import (
	"btaskee-quiz/internal/server"
	"errors"
)

func (s *QuizServer) getAuthUsername(ctx *server.Context) (string, error) {
	connCtx, ok := ctx.Conn().Context().(*server.ConnContext)
	if !ok || connCtx == nil {
		return "", errors.New("unauthorized")
	}
	return connCtx.Username(), nil
}
