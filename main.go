package main

import (
  "github.com/gofiber/fiber/v2"
  "github.com/gofiber/fiber/v2/middleware/logger"
  "log"
  "autogpt-api/handlers"
  "fmt"
)

func main() {

  app := fiber.New(fiber.Config{
    Prefork: true,
    CaseSensitive: true,
    ServerHeader: "Fiber",
    AppName: "autoGPT API v1.0.1",
  })
  
  if !fiber.IsChild() {
    fmt.Println("I'm the parent process")
  } else {
    fmt.Println("I'm a child process")
  }

  // Middleware to log requests
  app.Use(logger.New())

  // Routes
  app.Get("/", func(c *fiber.Ctx) error {
    return c.SendString("Hello World")
  })
  
  app.Get("/brain", handlers.OpenAIBrain)
  app.Post("/openai", handlers.OpenAIHandler)
  app.Post("/google", handlers.GoogleHandler)
  app.Post("/anthropic", handlers.AnthropicHandler)
  app.Get("/whisper", handlers.WhisperHandler)

  // Init server
  port := "8080"
  err := app.Listen(":" + port)

  if err != nil {
    log.Fatal("Can't connect to port", port, ": ", err)
  }
}
