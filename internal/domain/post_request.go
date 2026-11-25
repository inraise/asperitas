package domain

type PostRequest struct {
	Title    string  `json:"title" validate:"required,max=100"`
	Category string  `json:"category" validate:"required"`
	Type     string  `json:"type" validate:"required,oneof=text link"`
	Text     *string `json:"text,omitempty"`
	Url      *string `json:"url,omitempty"`
}
