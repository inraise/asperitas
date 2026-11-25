package domain

type Vote struct {
	User string `json:"user"`
	Vote int    `json:"vote"`
}
