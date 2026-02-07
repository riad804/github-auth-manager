package models

// Account represents non-sensitive account metadata
type Account struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Email    string `json:"email"`
}
