package config

import (
	"flag"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	loggerDebug = true

	botURL     = "https://example.com:8443"
	botMaxConn = 40
	botAutoDel = 20

	serverHost = "localhost"
	serverPort = "8443"

	webhookRetryCount = 5
	webhookRetrySleep = 2
	webhookDrop       = true

	tarantoolHost          = "localhost"
	tarantoolPort          = "3301"
	tarantoolUser          = "admin"
	tarantoolPass          = "admin"
	tarantoolTimeout       = 2
	tarantoolReconnect     = 2
	tarantoolMaxReconnects = 3
)

type Config struct {
	Logger struct {
		Debug bool `yaml:"debug"`
	} `yaml:"logger"`
	Bot struct {
		AutoDelete int `yaml:"auto_delete"`
		WebHook    struct {
			URL                string `yaml:"url"`
			MaxConnections     int    `yaml:"max_connections"`
			RetryCount         int    `yaml:"retry_count"`
			RetrySleep         int    `yaml:"retry_sleep"`
			DropPendingUpdates bool   `yaml:"drop_pending_updates"`
		} `yaml:"webhook"`
	} `yaml:"bot"`
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
	Tarantool struct {
		Host          string `yaml:"host"`
		Port          string `yaml:"port"`
		User          string `yaml:"user"`
		Pass          string `yaml:"password"`
		Timeout       int    `yaml:"timeout"`
		Reconnect     int    `yaml:"reconnect"`
		MaxReconnects uint   `yaml:"max_reconnects"`
	} `yaml:"tarantool"`
}

func New() *Config {
	return &Config{
		Logger: struct {
			Debug bool `yaml:"debug"`
		}{
			Debug: loggerDebug,
		},
		Bot: struct {
			AutoDelete int `yaml:"auto_delete"`
			WebHook    struct {
				URL                string `yaml:"url"`
				MaxConnections     int    `yaml:"max_connections"`
				RetryCount         int    `yaml:"retry_count"`
				RetrySleep         int    `yaml:"retry_sleep"`
				DropPendingUpdates bool   `yaml:"drop_pending_updates"`
			} `yaml:"webhook"`
		}{
			AutoDelete: botAutoDel,
			WebHook: struct {
				URL                string `yaml:"url"`
				MaxConnections     int    `yaml:"max_connections"`
				RetryCount         int    `yaml:"retry_count"`
				RetrySleep         int    `yaml:"retry_sleep"`
				DropPendingUpdates bool   `yaml:"drop_pending_updates"`
			}{
				URL:                botURL,
				MaxConnections:     botMaxConn,
				RetryCount:         webhookRetryCount,
				RetrySleep:         webhookRetrySleep,
				DropPendingUpdates: webhookDrop,
			},
		},
		Server: struct {
			Host string `yaml:"host"`
			Port string `yaml:"port"`
		}{
			Host: serverHost,
			Port: serverPort,
		},
		Tarantool: struct {
			Host          string `yaml:"host"`
			Port          string `yaml:"port"`
			User          string `yaml:"user"`
			Pass          string `yaml:"password"`
			Timeout       int    `yaml:"timeout"`
			Reconnect     int    `yaml:"reconnect"`
			MaxReconnects uint   `yaml:"max_reconnects"`
		}{
			Host:          tarantoolHost,
			Port:          tarantoolPort,
			User:          tarantoolUser,
			Pass:          tarantoolPass,
			Timeout:       tarantoolTimeout,
			Reconnect:     tarantoolReconnect,
			MaxReconnects: tarantoolMaxReconnects,
		},
	}
}

func (c *Config) Open(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	// Start YAML decoding from file
	if err = yaml.NewDecoder(file).Decode(&c); err != nil {
		return err
	}

	return nil
}

func PathFlag(path *string) {
	flag.StringVar(path, "config", "./configs/config.yaml", "path to config file")
}
