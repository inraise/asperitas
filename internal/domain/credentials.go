package domain

type Credentials struct {
	Username string `json:"username" validate:"required,username"`
	Password string `json:"password" validate:"required,min=8,maxbytes=72"`
}
