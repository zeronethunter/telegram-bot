package logger

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	Logrus *logrus.Logger
}

// New settings of logger.
func New() *Logger {
	newLogger := Logger{Logrus: logrus.New()}
	newLogger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		DisableQuote:    true,
		DisableColors:   true,
	})
	return &newLogger
}

var (
	lock         sync.Mutex
	SingleLogger *Logger
)

func GetInstance() *Logger {
	lock.Lock()
	defer lock.Unlock()

	if SingleLogger == nil {
		SingleLogger = New()
	}
	return SingleLogger
}

func ToLevel(level string) log.Lvl {
	switch level {
	case "debug":
		return log.DEBUG
	case "info":
		return log.INFO
	case "warn":
		return log.WARN
	case "error":
		return log.ERROR
	default:
		return log.INFO
	}
}

func toLogrusLevel(level log.Lvl) logrus.Level {
	switch level {
	case log.DEBUG:
		return logrus.DebugLevel
	case log.INFO:
		return logrus.InfoLevel
	case log.WARN:
		return logrus.WarnLevel
	case log.ERROR:
		return logrus.ErrorLevel
	}

	return logrus.InfoLevel
}

func toEchoLevel(level logrus.Level) log.Lvl {
	switch level {
	case logrus.DebugLevel:
		return log.DEBUG
	case logrus.InfoLevel:
		return log.INFO
	case logrus.WarnLevel:
		return log.WARN
	case logrus.ErrorLevel:
		return log.ERROR
	}

	return log.OFF
}

func (l *Logger) Println(v ...interface{}) {
	l.Logrus.Println(v...)
}

func (l *Logger) Output() io.Writer {
	return l.Logrus.Out
}

func (l *Logger) SetOutput(w io.Writer) {
	l.Logrus.SetOutput(w)
}

func (l *Logger) Level() log.Lvl {
	return toEchoLevel(l.Logrus.Level)
}

func (l *Logger) SetLevel(v log.Lvl) {
	l.Logrus.Level = toLogrusLevel(v)
}

func (l *Logger) SetHeader(_ string) {}

func (l *Logger) Formatter() logrus.Formatter {
	return l.Logrus.Formatter
}

func (l *Logger) SetFormatter(formatter logrus.Formatter) {
	l.Logrus.Formatter = formatter
}

func (l *Logger) Prefix() string {
	return ""
}

func (l *Logger) SetPrefix(_ string) {}

func (l *Logger) Print(i ...interface{}) {
	l.Logrus.Print(i...)
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.Logrus.Printf(format, args...)
}

func (l *Logger) Printj(j log.JSON) {
	b, _ := json.Marshal(j)
	l.Logrus.Println(string(b))
}

func (l *Logger) Debug(i ...interface{}) {
	l.Logrus.Debug(i...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Logrus.Debugf(format, args...)
}

func (l *Logger) Debugj(j log.JSON) {
	b, _ := json.Marshal(j)
	l.Logrus.Debugln(string(b))
}

func (l *Logger) Info(i ...interface{}) {
	l.Logrus.Info(i...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Logrus.Infof(format, args...)
}

func (l *Logger) Infoj(j log.JSON) {
	b, _ := json.Marshal(j)
	l.Logrus.Infoln(string(b))
}

func (l *Logger) Warn(i ...interface{}) {
	l.Logrus.Warn(i...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Logrus.Warnf(format, args...)
}

func (l *Logger) Warnj(j log.JSON) {
	b, _ := json.Marshal(j)
	l.Logrus.Warnln(string(b))
}

func (l *Logger) Error(i ...interface{}) {
	l.Logrus.Error(i...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Logrus.Errorf(format, args...)
}

func (l *Logger) Errorj(j log.JSON) {
	b, _ := json.Marshal(j)
	l.Logrus.Errorln(string(b))
}

func (l *Logger) Fatal(i ...interface{}) {
	l.Logrus.Fatal(i...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Logrus.Fatalf(format, args...)
}

func (l *Logger) Fatalj(j log.JSON) {
	b, _ := json.Marshal(j)
	l.Logrus.Fatalln(string(b))
}

func (l *Logger) Panic(i ...interface{}) {
	l.Logrus.Panic(i...)
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	l.Logrus.Panicf(format, args...)
}

func (l *Logger) Panicj(j log.JSON) {
	b, _ := json.Marshal(j)
	l.Logrus.Panicln(string(b))
}

func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()
			err := next(c)
			if err != nil {
				// Error message to telegram
				chatID, ok := c.Get("chatID").(int64)
				if ok {
					tgBot, ok := c.Get("bot").(*tgbotapi.BotAPI)
					if ok {
						msg := tgbotapi.NewMessage(chatID, "Sorry, I can't handle your request\nTry again later \xE2\x9B\x94")

						tgBot.Send(msg)
					}
				}

				// OK response to telegram server to prevent spam
				c.JSON(http.StatusOK, "Sorry, I can't handle your request\nTry again later \xE2\x9B\x94")
			}
			stop := time.Now()

			p := req.URL.Path

			bytesIn := req.Header.Get(echo.HeaderContentLength)

			if err != nil {
				GetInstance().Logrus.WithFields(logrus.Fields{
					"error":         err.Error(),
					"remote_ip":     c.RealIP(),
					"host":          req.Host,
					"uri":           req.RequestURI,
					"method":        req.Method,
					"path":          p,
					"referer":       req.Referer(),
					"user_agent":    req.UserAgent(),
					"status":        res.Status,
					"latency":       strconv.FormatInt(stop.Sub(start).Nanoseconds()/1000, 10),
					"latency_human": stop.Sub(start).String(),
					"bytes_in":      bytesIn,
					"bytes_out":     strconv.FormatInt(res.Size, 10),
				}).Error("ERROR REQUEST")

				return err
			}
			GetInstance().Logrus.WithFields(map[string]interface{}{
				"remote_ip":     c.RealIP(),
				"host":          req.Host,
				"uri":           req.RequestURI,
				"method":        req.Method,
				"path":          p,
				"referer":       req.Referer(),
				"user_agent":    req.UserAgent(),
				"status":        res.Status,
				"latency":       strconv.FormatInt(stop.Sub(start).Nanoseconds()/1000, 10),
				"latency_human": stop.Sub(start).String(),
				"bytes_in":      bytesIn,
				"bytes_out":     strconv.FormatInt(res.Size, 10),
			}).Debug("REQUEST")

			return nil
		}
	}
}
