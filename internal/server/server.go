package server

import (
	"time"

	"telegram-bot/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tarantool/go-tarantool"

	"telegram-bot/internal/bot"
	config "telegram-bot/internal/configuration"
	middlewareBot "telegram-bot/internal/middleware"
	passwdHandler "telegram-bot/internal/passwd/delivery"
	passwdRepository "telegram-bot/internal/passwd/repository"
	passwdUsecase "telegram-bot/internal/passwd/usecase"
)

type Server struct {
	Echo   *echo.Echo
	Bot    *bot.Bot
	Config *config.Config

	passwdHandler *passwdHandler.Handler
}

func New(cfg *config.Config) *Server {
	e := echo.New()
	e.Logger = logger.GetInstance()
	e.Debug = cfg.Logger.Debug

	e.Use(logger.Middleware())
	e.Use(middlewareBot.TokenCheck())
	e.Use(middleware.Secure())

	return &Server{
		Echo:   e,
		Config: cfg,
	}
}

func (s *Server) Start(botChan chan *bot.Bot) error {
	go func() {
		for createdBot := range botChan {
			s.Bot = createdBot
		}

		err := s.MakePasswd()
		if err != nil {
			logger.GetInstance().Fatalf("failed to make passwd service: %s", err)
		}
		s.MakeRoute()
	}()

	return s.Echo.Start(
		s.Config.Server.Host + ":" + s.Config.Server.Port,
	)
}

func (s *Server) MakeRoute() {
	s.Echo.Pre(middleware.RemoveTrailingSlash())

	s.Echo.POST("", s.passwdHandler.GetMessage)
}

func (s *Server) MakePasswd() error {
	opts := tarantool.Opts{
		Timeout:       time.Duration(s.Config.Tarantool.Timeout) * time.Second,
		Reconnect:     time.Duration(s.Config.Tarantool.Reconnect) * time.Second,
		MaxReconnects: s.Config.Tarantool.MaxReconnects,
		User:          s.Config.Tarantool.User,
		Pass:          s.Config.Tarantool.Pass,
	}
	t, err := passwdRepository.NewTarantool(s.Config.Tarantool.Host, s.Config.Tarantool.Port, opts)
	if err != nil {
		return err
	}
	s.passwdHandler = passwdHandler.NewHandler(passwdUsecase.NewPasswdUsecase(t), s.Bot)

	return nil
}
