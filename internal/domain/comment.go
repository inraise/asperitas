package domain

type Comment struct {
	Id      string `json:"id"`
	Author  User   `json:"author"`
	Body    string `json:"body"`
	Created string `json:"created"`
}

type CommentRequest struct {
	Comment string `json:"comment" validate:"required,max=2000"`
}
