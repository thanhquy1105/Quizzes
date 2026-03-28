package token

import "time"

// IMaker is an interface for managing tokens
type IMaker interface {
	// CreateToken creates a new token for a specific user and duration
	CreateToken(uid string, duration time.Duration, tokenType TokenType) (string, *Payload, error)

	// VerifyToken checks if the token is valid or not
	VerifyToken(token string, tokenType TokenType) (*Payload, error)

	// VerifyTokenWithoutExpired checks if the token is valid, excluding expires
	VerifyTokenWithoutExpired(token string, tokenType TokenType) (*Payload, error)
}
