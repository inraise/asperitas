package domain

type JwtClaims struct {
	User User `json:"user" validate:"required"`
	Iat  int  `json:"iat" validate:"required"`
	Exp  int  `json:"exp" validate:"required"`
}
