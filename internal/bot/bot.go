package bot

import (
	"encoding/json"
	"strconv"
	"time"

	"telegram-bot/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/labstack/gommon/log"

	config "telegram-bot/internal/configuration"
	"telegram-bot/internal/telegram"
)

type Bot struct {
	BotAPI     *tgbotapi.BotAPI
	token      string
	EncryptKey string
	AutoDelete int
	logger     *logger.Logger
}

func New(botToken, secretToken, encryptKey string, cfg *config.Config) (*Bot, error) {
	// Get instance of logger
	newLogger := logger.GetInstance()
	if cfg.Logger.Debug {
		newLogger.SetLevel(log.DEBUG)
	}

	// Set telegram-bot-api logger
	err := tgbotapi.SetLogger(newLogger)
	if err != nil {
		return nil, err
	}

	// Creating new bot with provided token
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}
	bot.Debug = cfg.Logger.Debug

	// Create webhook
	wh, err := telegram.New(cfg.Bot.WebHook.URL)
	if err != nil {
		return nil, err
	}

	wh.MaxConnections = cfg.Bot.WebHook.MaxConnections
	wh.SecretToken = secretToken
	wh.DropPendingUpdates = cfg.Bot.WebHook.DropPendingUpdates

	params, err := wh.Params()
	if err != nil {
		return nil, err
	}

	go webhookSetup(bot, wh, params, cfg, newLogger)

	return &Bot{
		BotAPI:     bot,
		logger:     logger.GetInstance(),
		AutoDelete: cfg.Bot.AutoDelete,
		EncryptKey: encryptKey,
		token:      botToken,
	}, nil
}

func webhookSetup(bot *tgbotapi.BotAPI, wh telegram.Config, params tgbotapi.Params, cfg *config.Config, logger *logger.Logger) {
	bot.MakeRequest(wh.Method(), params)

	info, _ := bot.GetWebhookInfo()
	if info.URL == "" {
		for i := 0; i < cfg.Bot.WebHook.RetryCount; i++ {
			time.Sleep(time.Duration(cfg.Bot.WebHook.RetrySleep) * time.Second)

			// Send webhook
			bot.MakeRequest(wh.Method(), params)

			info, _ = bot.GetWebhookInfo()

			logger.Info("Telegram webhook info: ", info)

			if info.LastErrorDate != 0 && i == cfg.Bot.WebHook.RetryCount-1 {
				logger.Fatalf("Telegram webhook callback failed: %s", info.LastErrorMessage)
			} else if info.LastErrorDate == 0 {
				break
			}

			logger.Warnf("Telegram webhook callback failed: %s. Retrying...", info.LastErrorMessage)
		}
	}

	logger.Info("Telegram webhook set up successfully")
	jsonInfo, _ := json.MarshalIndent(info, "", "    ")
	logger.Info("Telegram webhook info: ", string(jsonInfo))
}

func MenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	// Create keyboard
	return tgbotapi.NewReplyKeyboard(
		[][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton("1"),
				tgbotapi.NewKeyboardButton("2"),
				tgbotapi.NewKeyboardButton("3"),
			},
			{
				tgbotapi.NewKeyboardButton("help"),
			},
		}...,
	)
}

func NiceTimerCredentials(chatID int64, messageID int, bot *Bot, serviceName, username, password string) {
	timer := bot.AutoDelete

	for i := 0; i < timer; i++ {
		msg := tgbotapi.NewEditMessageText(
			chatID,
			messageID,
			"Your new credentials for "+serviceName+":\n"+
				"Username: `"+username+"`\n"+
				"Password: `"+password+"`\n\n"+
				"This message will be deleted in "+strconv.Itoa(timer-i)+" seconds",
		)
		msg.ParseMode = "markdown"

		bot.BotAPI.Send(msg)

		time.Sleep(time.Second)
	}

	msg := tgbotapi.NewDeleteMessage(chatID, messageID)
	bot.BotAPI.Send(msg)
}
