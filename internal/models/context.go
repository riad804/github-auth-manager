package models

type Context struct {
	Name     string `json:"name"`
	Token    string `json:"token"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}
