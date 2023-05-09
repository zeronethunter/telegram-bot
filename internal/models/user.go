package models

type User struct {
	ID    uint64 `json:"user_id"`
	Token string `json:"token"`
}
