package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"libs/loggerfiber"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Caller().Logger()

	loggerMiddleware := loggerfiber.New(log.Logger)

	app := fiber.New()
	app.Use(loggerMiddleware)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/validate", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "validation error",
		})
	})

	log.Fatal().Err(app.Listen(":3000")).Send()
}
