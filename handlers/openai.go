package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"io/ioutil"
	"github.com/gofiber/fiber/v2"
)

type OAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OAIRequestBody struct {
	Model          string         `json:"model"`
	Messages       []OAIMessage   `json:"messages"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

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
	// Read request body
	var requestBody struct {
		Prompt     string `json:"prompt"`
		Model      string `json:"model"`
		OutputJSON *bool  `json:"output_JSON"`
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

	// Create openai' request body
	oaiRequestBody := OAIRequestBody{
		Model: requestBody.Model,
		Messages: []OAIMessage{
			{Role: "system", Content: "You are a helpful assistant designed to output JSON."},
			{Role: "user", Content: requestBody.Prompt},
		},
	}
	if outputJSON {
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

	// Return response
	return c.Status(statusCode).Send(response)
}

