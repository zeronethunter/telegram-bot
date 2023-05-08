package main

import (
	"flag"
	"os"

	"telegram-bot/internal/bot"
	config "telegram-bot/internal/configuration"
	"telegram-bot/internal/server"
	"telegram-bot/pkg/logger"
)

// flag: --config <path_of_config>
func main() {
	/*---------------------------logger---------------------------*/
	l := logger.GetInstance()

	/*----------------------------flag----------------------------*/
	var configPath string
	config.PathFlag(&configPath)
	flag.Parse()

	/*---------------------------config---------------------------*/
	cfg := config.New()
	if err := cfg.Open(configPath); err != nil {
		l.Fatalf("failed to open config: %s", err)
	}

	botChan := make(chan *bot.Bot)

	/*----------------------------bot-----------------------------*/
	go func() {
		tgbot, err := bot.New(os.Getenv("BOT_TOKEN"), os.Getenv("WEBHOOK_SECRET_TOKEN"), os.Getenv("AES_KEY"), cfg)
		if err != nil {
			l.Fatalf("failed to create bot: %s", err)
		}

		botChan <- tgbot
		close(botChan)
	}()

	/*---------------------------server---------------------------*/
	s := server.New(cfg)

	/*---------------------------start----------------------------*/
	if err := s.Start(botChan); err != nil {
		l.Fatalf("failed to start server: %s", err)
	}
}
