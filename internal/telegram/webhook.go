package telegram

import (
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
	URL                *url.URL
	MaxConnections     int
	SecretToken        string
	DropPendingUpdates bool
}

func New(link string) (Config, error) {
	u, err := url.Parse(link)
	if err != nil {
		return Config{}, err
	}

	return Config{
		URL: u,
	}, nil
}

func (config Config) Method() string {
	return "setWebhook"
}

func (config Config) Params() (tgbotapi.Params, error) {
	params := make(tgbotapi.Params)

	if config.URL != nil {
		params["url"] = config.URL.String()
	}

	params.AddNonZero("max_connections", config.MaxConnections)
	params.AddNonEmpty("secret_token", config.SecretToken)
	params.AddBool("drop_pending_updates", config.DropPendingUpdates)

	return params, nil
}
