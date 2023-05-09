package models

type Credentials struct {
	UserID       uint64 `json:"user_id"`
	Token        string `json:"token"`
	ServiceName  string `json:"service_name"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}
