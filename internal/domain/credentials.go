package domain

type Credentials struct {
	Username string `json:"username" validate:"required,min=1,max=32"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}
