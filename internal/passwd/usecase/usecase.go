package passwdUsecase

import (
	"telegram-bot/internal/models"
	passwdRepository "telegram-bot/internal/passwd/repository"
	"telegram-bot/pkg"
)

type PasswdUsecase interface {
	CreateUser(userID int64, token string, key string) error
	SetToken(userID int64, token string, key string) error
	UpdateToken(userID int64, token string, key string) error
	GetUser(userID int64, key string) (models.User, error)
	DeleteCredentialsByUser(userID int64) error
	SetService(userID int64, serviceName string) error
	SetUsername(userID int64, serviceName, username string) error
	SetPassword(userID int64, serviceName, password, key string) error
	Get(userID int64, serviceName, key string) (string, string, error)
	GetAllServices(userID int64) ([]string, error)
	Delete(userID int64, serviceName string) error
	SetState(userID int64, state string) error
	SetStateLastServer(userID int64, lastService string) error
	GetState(userID int64) (models.State, error)
}

type passwdUsecase struct {
	PasswdUsecase
	storage passwdRepository.Storage
}

func NewPasswdUsecase(storage passwdRepository.Storage) PasswdUsecase {
	return &passwdUsecase{
		storage: storage,
	}
}

func (u *passwdUsecase) CreateUser(userID int64, token string, key string) error {
	token, err := pkg.Encrypt(token, key)
	if err != nil {
		return err
	}

	return u.storage.CreateUser(userID, token)
}

func (u *passwdUsecase) SetToken(userID int64, token string, key string) error {
	token, err := pkg.Encrypt(token, key)
	if err != nil {
		return err
	}

	return u.storage.SetToken(userID, token)
}

func (u *passwdUsecase) UpdateToken(userID int64, token string, key string) error {
	token, err := pkg.Encrypt(token, key)
	if err != nil {
		return err
	}

	return u.storage.UpdateToken(userID, token)
}

func (u *passwdUsecase) GetUser(userID int64, key string) (models.User, error) {
	user, err := u.storage.GetUser(userID)
	if err != nil {
		return models.User{}, err
	}

	if user == (models.User{}) {
		return models.User{}, nil
	}

	if user.Token, err = pkg.Decrypt(user.Token, key); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (u *passwdUsecase) DeleteCredentialsByUser(userID int64) error {
	return u.storage.DeleteCredentialsByUser(userID)
}

func (u *passwdUsecase) SetService(userID int64, serviceName string) error {
	return u.storage.SetService(userID, serviceName)
}

func (u *passwdUsecase) SetUsername(userID int64, serviceName, username string) error {
	return u.storage.SetUsername(userID, serviceName, username)
}

func (u *passwdUsecase) SetPassword(userID int64, serviceName, password, key string) error {
	password, err := pkg.Encrypt(password, key)
	if err != nil {
		return err
	}

	return u.storage.SetPassword(userID, serviceName, password)
}

func (u *passwdUsecase) Get(userID int64, serviceName, key string) (string, string, error) {
	data, err := u.storage.Get(userID, serviceName)
	if err != nil {
		return "", "", err
	}

	if data.PasswordHash, err = pkg.Decrypt(data.PasswordHash, key); err != nil {
		return "", "", err
	}

	return data.Username, data.PasswordHash, nil
}

func (u *passwdUsecase) GetAllServices(userID int64) ([]string, error) {
	data, err := u.storage.GetAllByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(data))
	for i, v := range data {
		result[i] = v.ServiceName
	}

	return result, nil
}

func (u *passwdUsecase) Delete(userID int64, serviceName string) error {
	return u.storage.Delete(userID, serviceName)
}

func (u *passwdUsecase) SetState(userID int64, state string) error {
	return u.storage.SetState(userID, state)
}

func (u *passwdUsecase) SetStateLastServer(userID int64, lastService string) error {
	return u.storage.SetStateLastServer(userID, lastService)
}

func (u *passwdUsecase) GetState(userID int64) (models.State, error) {
	data, err := u.storage.GetState(userID)
	if err != nil {
		return models.State{}, err
	}

	if data == (models.State{}) {
		return models.State{
			UserID: uint64(userID),
			State:  "default",
		}, nil
	}

	return data, nil
}

func (u *passwdUsecase) GetLastService(userID int64) (string, error) {
	data, err := u.storage.GetState(userID)
	if err != nil {
		return "", err
	}

	if data == (models.State{}) {
		return "", nil
	}

	return data.LastService, nil
}
