package handlers

import (
  "context"
  "time"
  "bytes"
  "encoding/json"
  "net/http"
  "os"
  "log"
  "io/ioutil"
  "fmt"
  "strings"

  "github.com/gofiber/fiber/v2"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
)


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
  var requestBody RequestBody
  if err := c.BodyParser(&requestBody); err != nil {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
      "error": "Cannot parse JSON",
    })
  }

  // Validate if the model belongs to google
  _, googleContents, err := processRequest(requestBody)
  if err != nil {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
      "error": err.Error(),
    })
  }

  // Create Google request body
  gRequestBody := GRequestBody{
    Model: requestBody.Model,
    Contents: googleContents,
  }

  if requestBody.OutputJSON == nil || *requestBody.OutputJSON {
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

  // If request is not successful, don't update the database
  if statusCode != http.StatusOK {
    return c.Status(statusCode).Send(response)
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

  // Get model prices from models.json
  inputPrice, outputPrice, err := getModelPrices(requestBody.Model, "google")
  if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "Error getting model prices",
    })
  }

  // Calculate usage
  inputTokens := googleResponse.UsageMetadata.PromptTokenCount
  outputTokens := googleResponse.UsageMetadata.CandidatesTokenCount
  inputUsage := float64(inputTokens) * (inputPrice / 1000000)
  outputUsage := float64(outputTokens) * (outputPrice / 1000000)

  // Update MongoDB
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  userID := requestBody.ID
  filter := bson.M{"id_user": userID}

  // Check if the user exists
  var user User
  err = userCollection.FindOne(ctx, filter).Decode(&user)
  if err != nil {
    return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
      "error": "User not found",
    })
  }

  // Ensure the model name with dot notation is handled properly
  modelKey := fmt.Sprintf("google.models.%s", strings.Replace(requestBody.Model, ".", "\u2024", -1))

  update := bson.M{
    "$inc": bson.M{
      "input_usage": inputUsage,
      "output_usage": outputUsage,
      "google.input_usage": inputUsage,
      "google.output_usage": outputUsage,
      modelKey + ".input_tokens": inputTokens,
      modelKey + ".output_tokens": outputTokens,
      modelKey + ".input_usage": inputUsage,
      modelKey + ".output_usage": outputUsage,
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
  opts := options.Update().SetUpsert(false)
  _, err = userCollection.UpdateOne(ctx, filter, update, opts)
  if err != nil {
    log.Printf("%v: ", err)
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "Error updating MongoDB",
    })
  }

  // Return response
  return c.Status(statusCode).Send(response)
}
