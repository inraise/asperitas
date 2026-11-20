package domain

type User struct {
	Id       string `json:"id" validate:"required,min=1"`
	Username string `json:"username" validate:"required,min=1,max=32"`
	Admin    bool   `json:"admin"`
}
