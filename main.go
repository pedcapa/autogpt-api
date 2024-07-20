package main

import (
  "github.com/gofiber/fiber/v2"
  "github.com/gofiber/fiber/v2/middleware/logger"
  "log"
  "autogpt-api/handlers"
)

func main() {

  app := fiber.New()

  // Middleware to log requests
  app.Use(logger.New())

  // Routes
  app.Get("/", func(c *fiber.Ctx) error {
    return c.SendString("Hello World")
  })

  app.Get("/openai", handlers.OpenAIHandler)
  app.Get("/gemini", handlers.GeminiHandler)
  app.Get("/claude", handlers.ClaudeHandler)
  app.Get("/whisper", handlers.WhisperHandler)

  // Init server
  port := "8080"
  err := app.Listen(":" + port)

  if err != nil {
    log.Fatal("Can't connect to port", port, ": ", err)
  }
}
