package passwdRepository

import (
	"fmt"

	"github.com/tarantool/go-tarantool"

	"telegram-bot/internal/models"
)

const maxServices = 50

type Storage interface {
	SetService(userID int64, serviceName string) error
	SetUsername(userID int64, serviceName string, username string) error
	SetPassword(userID int64, serviceName string, password string) error
	Get(userID int64, serviceName string) (models.Credentials, error)
	GetAllByUserID(userID int64) ([]models.Credentials, error)
	Delete(userID int64, serviceName string) error
	SetState(userID int64, state string) error
	SetStateLastServer(userID int64, lastService string) error
	GetState(userID int64) (models.State, error)
}

func parseCredential(data []interface{}) models.Credentials {
	if data == nil {
		return models.Credentials{}
	}

	if len(data) == 2 {
		return models.Credentials{
			UserID:      data[0].(uint64),
			ServiceName: data[1].(string),
		}
	} else if len(data) == 3 {
		return models.Credentials{
			UserID:      data[0].(uint64),
			ServiceName: data[1].(string),
			Username:    data[2].(string),
		}
	}

	return models.Credentials{
		UserID:       data[0].(uint64),
		ServiceName:  data[1].(string),
		Username:     data[2].(string),
		PasswordHash: data[3].(string),
	}
}

func parseCredentials(resp *tarantool.Response) []models.Credentials {
	var result []models.Credentials

	for _, data := range resp.Data {
		convertedData, ok := data.([]interface{})
		if ok {
			result = append(result, parseCredential(convertedData))
		}
	}

	return result
}

func parseState(data []interface{}) models.State {
	if data == nil {
		return models.State{}
	}

	if len(data) == 2 {
		return models.State{
			UserID: data[0].(uint64),
			State:  data[1].(string),
		}
	}

	return models.State{
		UserID:      data[0].(uint64),
		State:       data[1].(string),
		LastService: data[2].(string),
	}
}

type Tarantool struct {
	Storage
	conn *tarantool.Connection
}

func NewTarantool(host, port string, opts tarantool.Opts) (*Tarantool, error) {
	conn, err := tarantool.Connect(fmt.Sprintf("%s:%s", host, port), opts)
	if err != nil {
		return nil, err
	}

	if _, err = conn.Ping(); err != nil {
		return nil, err
	}

	return &Tarantool{
		conn: conn,
	}, nil
}

func (t *Tarantool) Close() error {
	return t.conn.Close()
}

func (t *Tarantool) SetService(userID int64, serviceName string) error {
	_, err := t.conn.Upsert(
		"credentials",
		[]interface{}{
			userID,
			serviceName,
		},
		[]interface{}{},
	)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tarantool) SetUsername(userID int64, serviceName, username string) error {
	_, err := t.conn.Upsert(
		"credentials",
		[]interface{}{
			userID,
			serviceName,
			username,
		},
		[]interface{}{
			[]interface{}{"=", 2, username},
		})
	if err != nil {
		return err
	}

	return nil
}

func (t *Tarantool) SetPassword(userID int64, serviceName, password string) error {
	_, err := t.conn.Upsert(
		"credentials",
		[]interface{}{
			userID,
			serviceName,
			password,
		},
		[]interface{}{
			[]interface{}{"=", 3, password},
		})
	if err != nil {
		return err
	}

	return nil
}

func (t *Tarantool) Get(userID int64, serviceName string) (models.Credentials, error) {
	resp, err := t.conn.Select("credentials", "primary", 0, 1, tarantool.IterEq, []interface{}{userID, serviceName})
	if err != nil {
		return models.Credentials{}, err
	}

	if len(resp.Data) == 0 || resp.Data == nil {
		return models.Credentials{}, nil
	}

	return parseCredential(resp.Data[0].([]interface{})), nil
}

func (t *Tarantool) GetAllByUserID(userID int64) ([]models.Credentials, error) {
	resp, err := t.conn.Select("credentials", "primary", 0, maxServices, tarantool.IterEq, []interface{}{userID})
	if err != nil {
		return nil, err
	}

	return parseCredentials(resp), nil
}

func (t *Tarantool) Delete(userID int64, serviceName string) error {
	_, err := t.conn.Delete("credentials", "primary", []interface{}{userID, serviceName})
	if err != nil {
		return err
	}

	return nil
}

func (t *Tarantool) SetState(userID int64, state string) error {
	_, err := t.conn.Upsert(
		"state",
		[]interface{}{
			userID,
			state,
		},
		[]interface{}{
			[]interface{}{"=", 1, state},
		})
	if err != nil {
		return err
	}

	return nil
}

func (t *Tarantool) SetStateLastServer(userID int64, lastService string) error {
	_, err := t.conn.Upsert(
		"state",
		[]interface{}{
			userID,
			lastService,
		},
		[]interface{}{
			[]interface{}{"=", 2, lastService},
		})
	if err != nil {
		return err
	}

	return nil
}

func (t *Tarantool) GetState(userID int64) (models.State, error) {
	resp, err := t.conn.Select("state", "primary", 0, 1, tarantool.IterEq, []interface{}{userID})
	if err != nil {
		return models.State{}, err
	}

	if len(resp.Data) == 0 || resp.Data == nil {
		return models.State{
			UserID: uint64(userID),
			State:  "default",
		}, nil
	}

	return parseState(resp.Data[0].([]interface{})), nil
}
