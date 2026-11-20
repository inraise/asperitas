package domain

type Comment struct {
	Id      string `json:"id" validate:"required,min=1"`
	Author  User   `json:"author" validate:"required"`
	Body    string `json:"body" validate:"required,min=1,max=2000"`
	Created string `json:"created" validate:"required"`
}
