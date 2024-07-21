package handlers

import (
  "github.com/gofiber/fiber/v2"
  "bytes"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "os"
  "fmt"
)

type Part struct {
  Text string `json:"text"`
}

type Content struct {
  Role string `json:"role"`
  Parts []Part `json:"parts"`
}

type GRequestBody struct {
  Contents []Content `json:"contents"`
}

func GoogleHandler(c *fiber.Ctx) error {
  // Read request body
  var requestBody struct {
    Model string `json:"model"`
    ResponseMIMEType string `json:"responseMimeType"`
    Contents []Content `json:"contents"`
  }

  if err := c.BodyParser(&requestBody); err != nil {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "Cannot parse JSON",
    })
  }

  // Set the Google API key and endpoint
  GKey := os.Getenv("GEMINI_API_KEY")
  if GKey == "" {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "GEMINI_API_KEY is not set",
    })
  }
  url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", requestBody.Model, GKey)
  
  // Convert request body to JSON
  jsonBody, err := json.Marshal(GRequestBody{Contents: requestBody.Contents})
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
