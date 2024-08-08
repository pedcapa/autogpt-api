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
  "log"
  "strings"

  "github.com/gofiber/fiber/v2"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
)

func OpenAIResponseJSON(requestBody OAIRequestBody) ([]byte, int, error) {
  // Set openai API key and endpoint
  OAIKey := os.Getenv("OPENAI_API_KEY")
  if OAIKey == "" {
    return nil, http.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "OPENAI_API_KEY is not set")
  }
  url := "https://api.openai.com/v1/chat/completions"

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

  // Add headers
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Authorization", "Bearer "+OAIKey)

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

func OpenAIHandler(c *fiber.Ctx) error {
  // Read requestBody
  var requestBody RequestBody
  if err := c.BodyParser(&requestBody); err != nil {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
      "error": "Cannot parse JSON",
    })
  }
 
  // Validate if the model belongs to openai
  openAIMessages, _, err := processRequest(requestBody)
  if err != nil {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
      "error": err.Error(),
    })
  }

  // Create openai' request body
  oaiRequestBody := OAIRequestBody{
    Model: requestBody.Model,
    Messages: openAIMessages,
  }

  // Config default settings
  if requestBody.OutputJSON == nil || *requestBody.OutputJSON {
    oaiRequestBody.ResponseFormat = &ResponseFormat{
      Type: "json_object",
    }
  }
  
  // Make openai' request
  response, statusCode, err := OpenAIResponseJSON(oaiRequestBody)
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
  var openAIResponse struct {
    Usage struct {
      PromptTokens int `json:"prompt_tokens"`
      CompletionTokens int `json:"completion_tokens"`
      TotalTokens int `json:"total_tokens"`
    } `json:"usage"`
  }
  if err := json.Unmarshal(response, &openAIResponse); err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "Error parsing response JSON",
    })
  }

  // Get model prices from models.json
  inputPrice, outputPrice, err := getModelPrices(requestBody.Model, "openai")
  if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "Error getting model prices",
    })
  }

  // Calculate usage 
  inputTokens := openAIResponse.Usage.PromptTokens
  outputTokens := openAIResponse.Usage.CompletionTokens
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
  modelKey := fmt.Sprintf("openai.models.%s", strings.Replace(requestBody.Model, ".", "\u2024", -1))

  update := bson.M{
    "$inc": bson.M{
      "input_usage": inputUsage,
      "output_usage": outputUsage,
      "openai.input_usage": inputUsage,
      "openai.output_usage": outputUsage,
      modelKey + ".input_tokens": inputTokens,
      modelKey + ".output_tokens": outputTokens,
      modelKey + ".input_usage":  inputUsage,
      modelKey + ".output_usage": outputUsage,
    },
    "$push": bson.M{
      "history": History{
        Company: "openai",
        Model: requestBody.Model,
        InputTokens: inputTokens,
        OutputTokens: outputTokens,
        InputUsage: inputUsage,
        OutputUsage: outputUsage,
        Created: time.Now().Unix(),
      },
    },
  }
  opts := options.Update().SetUpsert(false)
  _, err = userCollection.UpdateOne(ctx, filter, update, opts)
  if err != nil {
    log.Printf("%v", err)
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
      "error": "Error updating MongoDB",
    })
  }

  // Return response
  return c.Status(statusCode).Send(response)
}

