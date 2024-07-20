package handlers

import (
  "github.com/gofiber/fiber/v2"
  "os"
  "fmt"
)

func OpenAIBrain(c *fiber.Ctx) error {
  OAIKey := os.Getenv("OPENAI_API_KEY")
  if OAIKey == "" {
    return c.SendString("OPENAI_API_KEY is not set")
  }
  return c.SendString(fmt.Sprintf("OPENAI_API_KEY: %s", OAIKey))
}
