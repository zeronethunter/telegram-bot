package models

type State struct {
	UserID      uint64 `json:"user_id"`
	State       string `json:"state"`
	LastService string `json:"last_service"`
}
