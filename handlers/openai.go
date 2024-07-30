package handlers

import (
  "context"
  "time"
  "bytes"
  "encoding/json"
  "net/http"
  "os"
  "io/ioutil"

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

  // Calculate usage -----MAKE CHANGES-----
  inputTokens := openAIResponse.Usage.PromptTokens
  outputTokens := openAIResponse.Usage.CompletionTokens
  inputUsage := float64(inputTokens) * 0.0005 // example *change*
  outputUsage := float64(outputTokens) * 0.00025 // example *change*

  // Update MongoDB
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  userID := requestBody.ID
  filter := bson.M{"id_user": userID}


  update := bson.M{
    "$inc": bson.M{
      "input_usage": inputUsage,
      "output_usage": outputUsage,
      "openai.input_usage": inputUsage,
      "openai.output_usage": outputUsage,
      "openai.models." + requestBody.Model + ".input_tokens": inputTokens,
      "openai.models." + requestBody.Model + ".output_tokens": outputTokens,
      "openai.models." + requestBody.Model + ".input_usage": inputUsage,
      "openai.models." + requestBody.Model + ".output_usage": outputUsage,
    },
    "$push": bson.M{
      "history": History{
        Company: "openai",
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

