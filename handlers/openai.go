package handlers

import (
  "github.com/gofiber/fiber/v2"
)

func OpenAIHandler(c *fiber.Ctx) error {
  return c.SendString("OpenAI Handler")
}
