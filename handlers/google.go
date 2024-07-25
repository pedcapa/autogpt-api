package handlers

import (
  "context"
  "time"
  "bytes"
  "encoding/json"
  "net/http"
  "os"
  "io/ioutil"
  "fmt"

  "github.com/gofiber/fiber/v2"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
)

type Part struct {
  Text string `json:"text"`
}

type Content struct {
  Role string `json:"role"`
  Parts []Part `json:"parts"`
}

type GRequestBody struct {
  Model string `json:"model"`
  Contents []Content `json:"contents"`
  GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
}

type GenerationConfig struct {
  ResponseMIMEType string `json:"response_mime_type"`
}

func GoogleResponseJSON(requestBody GRequestBody) ([]byte, int, error) {
  // Set the Google API key and endpoint
  GKey := os.Getenv("GEMINI_API_KEY")
  if GKey == "" {
    return nil, http.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "GEMINI_API_KEY is not set")
  }
  url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", requestBody.Model, GKey)

  // Convert request body to JSON
  jsonBody, err := json.Marshal(requestBody)
  if err != nil {
    return nil, http.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "Error marshalling JSON")
  }

  // Make HTTP request
  req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
  if err != nil {
    return nil, http.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "Error creating request")
  }
  
  // Add header
  req.Header.Set("Content-Type", "application/json")

  // Send request
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    return nil, http.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "Error sending request")
  }
  defer resp.Body.Close()

  // Read response
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, http.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "Error reading response")
  }

  return body, resp.StatusCode, nil
}



func GoogleHandler(c *fiber.Ctx) error {
  // Read request body
  var requestBody struct {
    ID string `json:"id_user"`
    Model string `json:"model"`
    Contents []Content `json:"contents"`
    OutputJSON *bool `json:"output_JSON"`
  }

  if err := c.BodyParser(&requestBody); err != nil {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
      "error": "Cannot parse JSON",
    })
  }

  // Config default settings
  outputJSON := true
  if requestBody.OutputJSON != nil {
    outputJSON = *requestBody.OutputJSON
  }

  // Create Google request body
  gRequestBody := GRequestBody{
    Model: requestBody.Model,
    Contents: requestBody.Contents,
  }
  if outputJSON {
    gRequestBody.GenerationConfig = &GenerationConfig{
      ResponseMIMEType: "application/json",
    }
  }

  // Make Google request
  response, statusCode, err := GoogleResponseJSON(gRequestBody)
  if err != nil {
    return c.Status(statusCode).JSON(fiber.Map{
      "error": err.Error(),
    })
  }

  // Parse response to get token usage
  var googleResponse struct {
    UsageMetadata struct {
      PromptTokenCount int `json:"promptTokenCount"`
      CandidatesTokenCount int `json:"candidatesTokenCount"`
      TotalTokenCount int `json:"totalTokenCount"`
    } `json:"usageMetadata"`
  }
  if err := json.Unmarshal(response, &googleResponse); err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "Error parsing response JSON",
    })
  }

  // Calculate usage  -----CHANGE PRICES-----
  inputTokens := googleResponse.UsageMetadata.PromptTokenCount
  outputTokens := googleResponse.UsageMetadata.CandidatesTokenCount
  inputUsage := float64(inputTokens) * 0.00035  // change
  outputUsage := float64(outputTokens) * 0.00105  // change

  // Update MongoDB
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  userID := requestBody.ID
  filter := bson.M{"id_user": userID}
  update := bson.M{
    "$inc": bson.M{
      "input_usage": inputUsage,
      "output_usage": outputUsage,
      "google.input_usage": inputUsage,
      "google.output_usage": outputUsage,
      "google.models." + requestBody.Model + ".input_tokens": inputTokens,
      "google.models." + requestBody.Model + ".output_tokens": outputTokens,
      "google.models." + requestBody.Model + ".input_usage": inputUsage,
      "google.models." + requestBody.Model + ".output_usage": outputUsage,
    },
    "$push": bson.M{
      "history": History{
        Company: "google",
        Model: requestBody.Model,
        InputTokens: inputTokens,
        OutputTokens: outputTokens,
        InputUsage: inputUsage,
        OutputUsage: outputUsage,
        Created: time.Now(),
      },
    },
  }
  opts := options.Update().SetUpsert(true)
  _, err = userCollection.UpdateOne(ctx, filter, update, opts)
  if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "Error updating MongoDB",
    })
  }

  // Return response
  return c.Status(statusCode).Send(response)
}
