package handlers

import (
  "github.com/gofiber/fiber/v2"
  "bytes"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "os"
)

type ANTMessage struct {
  Role string `json:"role"`
  Content string `json:"content"`
}

type ANTRequestBody struct {
  Model string `json:"model"`
  MaxTokens int `json:"max_tokens"`
  Messages []ANTMessage `json:"messages"`
}

func AnthropicHandler(c *fiber.Ctx) error {
  // Read request body
  var requestBody ANTRequestBody
  if err := c.BodyParser(&requestBody); err != nil {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
      "error": "Cannot parse JSON",
    })
  }

  // Set the Anthropic API key and endpoint
  ANTKey := os.Getenv("CLAUDE_API_KEY")
  if ANTKey == "" {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "CLAUDE_API_KEY is not set",
    })
  }
  url := "https://api.anthropic.com/v1/messages"

  // Convert request body to JSON
  jsonBody, err := json.Marshal(requestBody)
  if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "Error marshalling JSON",
    })
  }

  // Make HTTP request
  req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
  if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "Error creating request",
    })
  }

  // Add headers
  req.Header.Set("x-api-key", ANTKey)
  req.Header.Set("anthropic-version", "2023-06-01")
  req.Header.Set("Content-Type", "application/json")

  // Send request
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "Error sending request",
    })
  }
  defer resp.Body.Close()

  // Read response
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "Error reading response",
    })
  }

  // Return response client
  return c.Status(resp.StatusCode).Send(body)
}
