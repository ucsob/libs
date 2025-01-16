package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"libs/metricsfiber"
	"log"
)

func main() {
	app := fiber.New()

	registry := prometheus.DefaultRegisterer
	metricsUrl := "/metrics"
	mw := metricsfiber.New(metricsfiber.Config{
		Registry:  registry,
		Namespace: "prefix",
		SubSystem: "server",
		SkipPaths: []string{"/favicon.ico", metricsUrl},
	})
	app.Use(mw.Serve)

	app.Get(metricsUrl, adaptor.HTTPHandler(promhttp.HandlerFor(registry.(prometheus.Gatherer), promhttp.HandlerOpts{})))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/validate", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "validation error",
		})
	})

	log.Fatal(app.Listen(":3000"))
}
