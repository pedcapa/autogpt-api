package handlers

import (
  "github.com/gofiber/fiber/v2"
)

func WhisperHandler(c *fiber.Ctx) error {
  return c.SendString("Whisper Handler")
}
