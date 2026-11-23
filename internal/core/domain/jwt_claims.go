package domain

import "github.com/golang-jwt/jwt/v5"

type JwtClaims struct {
	User User `json:"user"`
	jwt.RegisteredClaims
}
