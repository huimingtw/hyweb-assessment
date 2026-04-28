package model

import "github.com/golang-jwt/jwt/v5"

// JWTClaims is shared by handler (signing) and middleware (parsing) to avoid import cycles.
type JWTClaims struct {
	Email   string `json:"email"`
	Updated string `json:"updated"` // RFC3339 UTC string of user.Updated
	jwt.RegisteredClaims
}
