package handlers

import (
  "github.com/gofiber/fiber/v2"
  "bytes"
  "encoding/json"
  "net/http"
  "os"
  "io/ioutil"
)

type OAIMessage struct {
  Role string `json:"role"`
  Content string `json:"content"`
}

type OAIRequestBody struct {
  Model string `json:"model"`
  Messages []OAIMessage `json:"messages"`
}

func OpenAIHandler(c *fiber.Ctx) error {
  // Read request body
  var requestBody OAIRequestBody
  if err := c.BodyParser(&requestBody); err != nil {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
      "error": "Cannot parse JSON",
    })
  }
  
  // Set the OAI API key and endpoint
  OAIKey := os.Getenv("OPENAI_API_KEY")
  if OAIKey == "" {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "OPENAI_API_KEY is not set",
    })
  }
  url := "https://api.openai.com/v1/chat/completions"
 
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
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Authorization", "Bearer " + OAIKey)

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

  // Return client response
  return c.Status(resp.StatusCode).Send(body)
}
