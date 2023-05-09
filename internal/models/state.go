package models

type State struct {
	UserID      uint64 `json:"user_id"`
	State       string `json:"state"`
	LastService string `json:"last_service"`
}

const (
	StateDefault            = "default"
	StateSetToken           = "setToken"
	StateCheckToken         = "checkToken"
	StateUpdateTokenConfirm = "updateTokenConfirm"
	StateUpdateTokenInput   = "updateTokenInput"
	StateUpdateToken        = "updateToken"
	StateSetService         = "setService"
	StateSetUsername        = "setUsername"
	StateSetPassword        = "setPassword"
	StateGetService         = "getService"
	StateDeleteService      = "deleteService"
)
