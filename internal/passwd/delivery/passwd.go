package passwdHandler

import (
	"encoding/json"
	"strconv"

	"telegram-bot/internal/models"

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

	user, err := h.usecase.GetUser(u.Message.From.ID, h.bot.EncryptKey)
	if err != nil {
		return err
	}

	state, err := h.usecase.GetState(u.Message.From.ID)
	if err != nil {
		return err
	}

	if user == (models.User{}) && state.State != models.StateSetToken {
		return h.start(u.Message)
	}

	if state.State == models.StateSetToken {
		switch u.Message.Text {
		case models.BackToMenuCMD, models.SetCMD, models.GetCMD, models.DelCMD, models.UpdateTokenCMD, models.HelpCMD:
			msg := tgbotapi.NewMessage(u.Message.Chat.ID, "Security password can't be a command, write it again")
			_, err = h.bot.BotAPI.Send(msg)
			return err
		}
	}

	if u.Message.Command() == "start" {
		if err = h.usecase.SetState(u.Message.From.ID, models.StateDefault); err != nil {
			return err
		}
		return h.startExisting(u.Message)
	}

	switch u.Message.Text {
	case models.HelpCMD:
		return h.help(u.Message)
	case models.SetCMD:
		return h.set(u.Message)
	case models.GetCMD:
		return h.askToken(u.Message)
	case models.DelCMD:
		return h.delete(u.Message)
	case models.UpdateTokenCMD:
		return h.updateTokenQ(u.Message)
	case models.BackToMenuCMD:
		if err = h.usecase.SetState(u.Message.From.ID, models.StateDefault); err != nil {
			return err
		}

		switch state.State {
		case models.StateSetUsername:
		case models.StateSetPassword:
			return h.usecase.Delete(u.Message.From.ID, state.LastService)
		}

		return h.help(u.Message)
	default:
		switch state.State {
		case models.StateCheckToken:
			return h.checkToken(u.Message, user.Token)
		case models.StateSetToken:
			return h.setToken(u.Message)

		case models.StateUpdateTokenConfirm:
			return h.updateTokenQ(u.Message)
		case models.StateUpdateTokenInput:
			return h.updateTokenInput(u.Message)
		case models.StateUpdateToken:
			return h.updateToken(u.Message)

		case models.StateSetService:
			return h.setService(u.Message)
		case models.StateSetUsername:
			return h.setUsername(u.Message, state.LastService)
		case models.StateSetPassword:
			return h.setPassword(u.Message, state.LastService)

		case models.StateGetService:
			return h.getService(u.Message)

		case models.StateDeleteService:
			return h.deleteService(u.Message)
		default:
			if err = h.usecase.SetState(u.Message.From.ID, models.StateDefault); err != nil {
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

func (h Handler) YesOrNoKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Yes"),
			tgbotapi.NewKeyboardButton("No"),
		),
	)
}

func (h Handler) help(m *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(
		m.Chat.ID,
		"1. Set credentials for service. \xF0\x9F\x94\x92\n2. Get login and password of service. \xF0\x9F\x94\x91\n3. Delete service. \xE2\x9D\x8C\n4. Change security password. \xF0\x9F\x94\x83\n\nEnter the number of the desired action:",
	)
	msg.ReplyMarkup = bot.MenuKeyboard()

	_, err := h.bot.BotAPI.Send(msg)

	return err
}

func (h Handler) startExisting(m *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(
		m.Chat.ID,
		"Hello again, ["+m.From.UserName+"](tg://user?id="+strconv.FormatInt(m.From.ID, 10)+")\\!\n\n"+
			"Enter what you want to do:",
	)
	msg.ReplyMarkup = bot.MenuKeyboard()
	msg.ParseMode = "MarkdownV2"

	if _, err := h.bot.BotAPI.Send(msg); err != nil {
		return err
	}

	return nil
}

func (h Handler) start(m *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(
		m.Chat.ID,
		"Hello, ["+m.From.UserName+"](tg://user?id="+strconv.FormatInt(m.From.ID, 10)+")\\!\n"+
			"I'm a password manager bot\\.\n"+
			"For security purposes, I will ask you to enter your security password\\.\n"+
			"*If it is lost, all data will be deleted*\\. Be careful\\!\\.\n\n"+
			"Enter your security password:",
	)
	msg.ParseMode = "MarkdownV2"

	if _, err := h.bot.BotAPI.Send(msg); err != nil {
		return err
	}

	return h.usecase.SetState(m.From.ID, models.StateSetToken)
}

func (h Handler) askToken(m *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(m.Chat.ID, "Enter security password:")
	msg.ReplyMarkup = h.BackToMenuKeyboard()

	if _, err := h.bot.BotAPI.Send(msg); err != nil {
		return err
	}

	return h.usecase.SetState(m.From.ID, "checkToken")
}

func (h Handler) checkToken(m *tgbotapi.Message, realToken string) error {
	if m.Text != realToken {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Wrong security password\\.\nTry again:")
		msg.ReplyMarkup = h.BackToMenuKeyboard()

		if _, err := h.bot.BotAPI.Send(msg); err != nil {
			return err
		}

		return nil
	}

	return h.get(m)
}

func (h Handler) updateTokenQ(m *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(m.Chat.ID, "This will delete all your passwords.\nAre you sure?")
	msg.ReplyMarkup = h.YesOrNoKeyboard()

	if _, err := h.bot.BotAPI.Send(msg); err != nil {
		return err
	}

	return h.usecase.SetState(m.From.ID, models.StateUpdateTokenInput)
}

func (h Handler) updateTokenInput(m *tgbotapi.Message) error {
	if m.Text == "Yes" {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Enter new security password:")
		msg.ReplyMarkup = h.BackToMenuKeyboard()

		if _, err := h.bot.BotAPI.Send(msg); err != nil {
			return err
		}

		return h.usecase.SetState(m.From.ID, models.StateUpdateToken)
	}

	err := h.usecase.SetState(m.From.ID, models.StateDefault)
	if err != nil {
		return err
	}

	return h.help(m)
}

func (h Handler) setToken(m *tgbotapi.Message) error {
	err := h.usecase.SetToken(m.From.ID, m.Text, h.bot.EncryptKey)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, "Security password saved successfully! \xE2\x9C\x85")
	msg.ReplyMarkup = bot.MenuKeyboard()

	if _, err = h.bot.BotAPI.Send(msg); err != nil {
		return err
	}

	return h.usecase.SetState(m.From.ID, models.StateDefault)
}

func (h Handler) updateToken(m *tgbotapi.Message) error {
	err := h.usecase.UpdateToken(m.From.ID, m.Text, h.bot.EncryptKey)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, "Security password updated successfully! \xE2\x9C\x85")
	msg.ReplyMarkup = bot.MenuKeyboard()

	if _, err = h.bot.BotAPI.Send(msg); err != nil {
		return err
	}

	return h.usecase.SetState(m.From.ID, models.StateDefault)
}

func (h Handler) set(m *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(m.Chat.ID, "Enter service:")
	msg.ReplyMarkup = h.BackToMenuKeyboard()

	if _, err := h.bot.BotAPI.Send(msg); err != nil {
		return err
	}

	return h.usecase.SetState(m.From.ID, models.StateSetService)
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

	return h.usecase.SetState(m.From.ID, models.StateSetUsername)
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

	return h.usecase.SetState(m.From.ID, models.StateSetPassword)
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

	return h.usecase.SetState(m.From.ID, models.StateDefault)
}

func (h Handler) get(m *tgbotapi.Message) error {
	var err error

	msg := tgbotapi.NewMessage(m.Chat.ID, "Correct \xE2\x9C\x85\nEnter service:")
	if msg.ReplyMarkup, err = h.allServicesKeyboard(m.From.ID); err != nil {
		return err
	}

	if err = h.usecase.SetState(m.From.ID, models.StateGetService); err != nil {
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

		return h.usecase.SetState(m.From.ID, models.StateDefault)
	}

	msg := tgbotapi.NewMessage(
		m.Chat.ID,
		"Your credentials for "+m.Text+":\n"+
			"Username: `"+username+"`\n"+
			"Password: `"+password+"`\n\n",
	)
	msg.ParseMode = "markdown"

	response, err := h.bot.BotAPI.Send(msg)
	if err != nil {
		return err
	}

	go bot.NiceTimerCredentials(response.Chat.ID, response.MessageID, h.bot, m.Text, username, password)

	return h.usecase.SetState(m.From.ID, models.StateDefault)
}

func (h Handler) delete(m *tgbotapi.Message) error {
	var err error

	msg := tgbotapi.NewMessage(m.Chat.ID, "Enter service:")
	if msg.ReplyMarkup, err = h.allServicesKeyboard(m.From.ID); err != nil {
		return err
	}

	if err = h.usecase.SetState(m.From.ID, models.StateDeleteService); err != nil {
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

	return h.usecase.SetState(m.From.ID, models.StateDefault)
}
