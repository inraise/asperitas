package domain

type PostRequest struct {
	Title    string  `json:"title" validate:"required,min=1,max=100"`
	Category string  `json:"category" validate:"required,min=1"`
	Type     string  `json:"type" validate:"required,oneof=text link"`
	Text     *string `json:"text,omitempty"`
	Url      *string `json:"url,omitempty"`
}
