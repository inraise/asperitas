package domain

type Vote struct {
	User string `json:"user" validate:"required,min=1"`
	Vote int    `json:"vote" validate:"required"`
}
