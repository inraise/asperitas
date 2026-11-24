package domain

type Post struct {
	Id       string `json:"id" validate:"required,min=1"`
	Title    string `json:"title" validate:"required,min=1,max=100"`
	Category string `json:"category" validate:"required,min=1"`
	Author   User   `json:"author" validate:"required"`
	Score    int    `json:"score" validate:"required"`

	Votes    []Vote    `json:"votes" validate:"required"`
	Comments []Comment `json:"comments" validate:"required"`

	Created          string `json:"created" validate:"required"`
	Views            int    `json:"views" validate:"required"`
	UpvotePercentage int    `json:"upvotePercentage" validate:"required,min=0,max=100"`

	Type string  `json:"type" validate:"required,oneof=text link"`
	Text *string `json:"text,omitempty"`
	Url  *string `json:"url,omitempty"`
}
