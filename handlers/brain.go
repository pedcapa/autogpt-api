package handlers

import (
  "github.com/gofiber/fiber/v2"
  "os"
  "fmt"
)

func OpenAIBrain(c *fiber.Ctx) error {
  OAIKey := os.Getenv("OPENAI_API_KEY")
  GKey := os.Getenv("GEMINI_API_KEY")
  ANTKey := os.Getenv("CLAUDE_API_KEY")

  response := "Still working on this route...\n"

  if OAIKey == "" {
    response += "OPENAI_API_KEY is not set"
  } else {
    response += fmt.Sprintf("OPENAI_API_KEY: %s", OAIKey)
  }
  if GKey == "" {
    response += "\nGEMINI_API_KEY is not set"
  } else {
    response += fmt.Sprintf("\nGEMINI_API_KEY: %s", GKey)
  }
  if ANTKey == "" {
    response += "\nCLAUDE_API_KEY is not set"
  } else {
    response += fmt.Sprintf("\nCLAUDE_API_KEY: %s", ANTKey)
  }
  
  return c.SendString(response)

}
