package handlers

import "github.com/gofiber/fiber/v2"

func ClaudeHandler(c *fiber.Ctx) error {
  return c.SendString("Claude Handler")
}
