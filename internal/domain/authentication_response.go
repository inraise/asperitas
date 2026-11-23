package domain

type AuthenticationResponse struct {
	Token string `json:"token" validate:"required,min=1"`
}
