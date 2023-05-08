package passwdHandler

import (
	"encoding/json"
	"strconv"

	"telegram-bot/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/labstack/echo/v4"

	"telegram-bot/internal/bot"
	passwdUsecase "telegram-bot/internal/passwd/usecase"
)

type Handler struct {
	usecase passwdUsecase.PasswdUsecase
	bot     *bot.Bot
	logger  *logger.Logger
}

func NewHandler(usecase passwdUsecase.PasswdUsecase, bot *bot.Bot) *Handler {
	return &Handler{
		usecase: usecase,
		bot:     bot,
		logger:  logger.GetInstance(),
	}
}

func (h Handler) GetMessage(c echo.Context) error {
	var u tgbotapi.Update

	err := c.Bind(&u)
	if err != nil {
		return err
	}

	toLog, _ := json.MarshalIndent(u, "", "  ")
	h.logger.Debugf("update: \n%s", toLog)

	c.Set("chatID", u.Message.Chat.ID)
	c.Set("bot", h.bot.BotAPI)

	if u.Message.Command() == "start" {
		return h.start(u.Message)
	}

	state, err := h.usecase.GetState(u.Message.From.ID)
	if err != nil {
		return err
	}

	switch u.Message.Text {
	case "help":
		return h.help(u.Message)
	case "1":
		return h.set(u.Message)
	case "2":
		return h.get(u.Message)
	case "3":
		return h.delete(u.Message)
	case "Back to menu <<":
		if err = h.usecase.SetState(u.Message.From.ID, "default"); err != nil {
			return err
		}

		switch state.State {
		case "setUsername":
		case "setPassword":
			return h.usecase.Delete(u.Message.From.ID, state.LastService)
		}

		return h.help(u.Message)
	default:
		switch state.State {
		case "setService":
			return h.setService(u.Message)
		case "setUsername":
			return h.setUsername(u.Message, state.LastService)
		case "setPassword":
			return h.setPassword(u.Message, state.LastService)
		case "getService":
			return h.getService(u.Message)
		case "deleteService":
			return h.deleteService(u.Message)
		default:
			if err = h.usecase.SetState(u.Message.From.ID, "default"); err != nil {
				return err
			}

			return h.help(u.Message)
		}
	}
}

func (h Handler) allServicesKeyboard(userID int64) (tgbotapi.ReplyKeyboardMarkup, error) {
	var keyboard [][]tgbotapi.KeyboardButton
	var row []tgbotapi.KeyboardButton

	services, err := h.usecase.GetAllServices(userID)
	if err != nil {
		return tgbotapi.ReplyKeyboardMarkup{}, err
	}

	for _, service := range services {
		row = append(row, tgbotapi.KeyboardButton{Text: service})

		if len(row)+1%5 == 0 {
			keyboard = append(keyboard, row)
			row = nil
		}
	}

	if len(row) > 0 {
		keyboard = append(keyboard, row)
	}

	keyboard = append(keyboard, []tgbotapi.KeyboardButton{{Text: "Back to menu <<"}})

	return tgbotapi.NewReplyKeyboard(keyboard...), nil
}

func (h Handler) BackToMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Back to menu <<"),
		),
	)
}

func (h Handler) help(m *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(
		m.Chat.ID,
		"1. Set credentials for service. \xF0\x9F\x94\x92\n2. Get login and password of service. \xF0\x9F\x94\x91\n3. Delete service. \xE2\x9D\x8C",
	)
	msg.ReplyMarkup = bot.MenuKeyboard()

	_, err := h.bot.BotAPI.Send(msg)

	return err
}

func (h Handler) start(m *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(
		m.Chat.ID,
		"Hello, ["+m.From.UserName+"](tg://user?id="+strconv.FormatInt(m.From.ID, 10)+")\\!\n"+
			"I'm a password manager bot\\.\n"+
			"Enter `help` command to add service credentials\\.",
	)
	msg.ReplyMarkup = bot.MenuKeyboard()
	msg.ParseMode = "MarkdownV2"

	if _, err := h.bot.BotAPI.Send(msg); err != nil {
		return err
	}

	return h.usecase.SetState(m.From.ID, "default")
}

func (h Handler) set(m *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(m.Chat.ID, "Enter service:")
	msg.ReplyMarkup = h.BackToMenuKeyboard()

	if _, err := h.bot.BotAPI.Send(msg); err != nil {
		return err
	}

	return h.usecase.SetState(m.From.ID, "setService")
}

func (h Handler) setService(m *tgbotapi.Message) error {
	if err := h.usecase.SetService(m.From.ID, m.Text); err != nil {
		return err
	}

	if err := h.usecase.SetStateLastServer(m.From.ID, m.Text); err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, "Enter username:")
	msg.ReplyMarkup = h.BackToMenuKeyboard()

	if _, err := h.bot.BotAPI.Send(msg); err != nil {
		return err
	}

	return h.usecase.SetState(m.From.ID, "setUsername")
}

func (h Handler) setUsername(m *tgbotapi.Message, lastService string) error {
	if err := h.usecase.SetUsername(m.From.ID, lastService, m.Text); err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, "Enter password:")
	msg.ReplyMarkup = h.BackToMenuKeyboard()

	if _, err := h.bot.BotAPI.Send(msg); err != nil {
		return err
	}

	return h.usecase.SetState(m.From.ID, "setPassword")
}

func (h Handler) setPassword(m *tgbotapi.Message, lastService string) error {
	err := h.usecase.SetPassword(m.From.ID, lastService, m.Text, h.bot.EncryptKey)
	if err != nil {
		return err
	}

	username := ""

	if username, _, err = h.usecase.Get(m.From.ID, lastService, h.bot.EncryptKey); err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(
		m.Chat.ID,
		`Successfully saved\! \xE2\x9C\x85\n`+
			"Your new credentials for "+lastService+":\n"+
			"Username: `"+username+"`\n"+
			"Password: `"+m.Text+"`",
	)
	msg.ParseMode = "markdown"
	msg.ReplyToMessageID = m.MessageID

	response, err := h.bot.BotAPI.Send(msg)
	if err != nil {
		return err
	}

	go bot.NiceTimerCredentials(response.Chat.ID, response.MessageID, h.bot, lastService, username, m.Text)

	return h.usecase.SetState(m.From.ID, "default")
}

func (h Handler) get(m *tgbotapi.Message) error {
	var err error

	msg := tgbotapi.NewMessage(m.Chat.ID, "Enter service:")
	if msg.ReplyMarkup, err = h.allServicesKeyboard(m.From.ID); err != nil {
		return err
	}

	if err = h.usecase.SetState(m.From.ID, "getService"); err != nil {
		return err
	}

	_, err = h.bot.BotAPI.Send(msg)

	return err
}

func (h Handler) getService(m *tgbotapi.Message) error {
	username, password, err := h.usecase.Get(m.From.ID, m.Text, h.bot.EncryptKey)
	if err != nil {
		return err
	}

	if username == "" || password == "" {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Service not found!")
		msg.ReplyMarkup = bot.MenuKeyboard()

		if _, err = h.bot.BotAPI.Send(msg); err != nil {
			return err
		}

		return h.usecase.SetState(m.From.ID, "default")
	}

	msg := tgbotapi.NewMessage(
		m.Chat.ID,
		"Your new credentials for "+m.Text+":\n"+
			"Username: `"+username+"`\n"+
			"Password: `"+password+"`\n\n",
	)
	msg.ParseMode = "markdown"

	response, err := h.bot.BotAPI.Send(msg)
	if err != nil {
		return err
	}

	go bot.NiceTimerCredentials(response.Chat.ID, response.MessageID, h.bot, m.Text, username, password)

	return h.usecase.SetState(m.From.ID, "default")
}

func (h Handler) delete(m *tgbotapi.Message) error {
	var err error

	msg := tgbotapi.NewMessage(m.Chat.ID, "Enter service:")
	if msg.ReplyMarkup, err = h.allServicesKeyboard(m.From.ID); err != nil {
		return err
	}

	if err = h.usecase.SetState(m.From.ID, "deleteService"); err != nil {
		return err
	}

	_, err = h.bot.BotAPI.Send(msg)

	return err
}

func (h Handler) deleteService(m *tgbotapi.Message) error {
	if err := h.usecase.Delete(m.From.ID, m.Text); err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, "Successfully deleted! \xE2\x9C\x85")
	msg.ReplyMarkup = bot.MenuKeyboard()

	if _, err := h.bot.BotAPI.Send(msg); err != nil {
		return err
	}

	return h.usecase.SetState(m.From.ID, "default")
}
