package loggerfiber

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

func New(l zerolog.Logger) fiber.Handler {
	var (
		once       sync.Once
		errHandler fiber.ErrorHandler
	)

	return func(c *fiber.Ctx) error {
		once.Do(func() {
			errHandler = c.App().ErrorHandler
		})

		start := time.Now()

		chainErr := c.Next()
		if chainErr != nil {
			if err := errHandler(c, chainErr); err != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		code := c.Response().StatusCode()

		log.Logger = l.With().
			Int("status", code).
			Str("method", c.Route().Method).
			Str("path", c.Route().Path).
			Str("ip", c.IP()).
			Str("latency", time.Since(start).String()).
			Str("user-agent", c.Get(fiber.HeaderUserAgent)).
			Logger()

		msgBuffer := new(bytes.Buffer)
		_ = json.Compact(msgBuffer, c.BodyRaw())
		switch {
		case code >= 200 && code < 300:
			log.Debug().Msg(msgBuffer.String())
		default:
			log.Error().Err(chainErr).Msg(msgBuffer.String())
		}

		return chainErr
	}
}
