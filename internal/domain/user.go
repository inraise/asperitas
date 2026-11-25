package domain

import "regexp"

type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Admin    bool   `json:"admin,omitempty"`
}

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func ValidUsername(s string) bool {
	return len(s) >= 1 && len(s) <= 32 && usernameRe.MatchString(s)
}
