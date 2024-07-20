package handlers

import (
  "github.com/gofiber/fiber/v2"
)

func GeminiHandler(c *fiber.Ctx) error {
  return c.SendString("Gemini Handler")
}
