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
	// Configurar el API key y la URL de OpenAI
	OAIKey := os.Getenv("OPENAI_API_KEY")
	if OAIKey == "" {
		return nil, http.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "OPENAI_API_KEY is not set")
	}
	url := "https://api.openai.com/v1/chat/completions"

	// Convertir el cuerpo de la solicitud a JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, http.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "Error marshalling JSON")
	}

	// Crear la solicitud HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, http.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "Error creating request")
	}

	// Añadir headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+OAIKey)

	// Enviar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "Error sending request")
	}
	defer resp.Body.Close()

	// Leer la respuesta
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "Error reading response")
	}

	return body, resp.StatusCode, nil
}

func OpenAIHandler(c *fiber.Ctx) error {
	// Leer datos del cuerpo de la solicitud
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

	// Configurar valores por defecto
	outputJSON := true
	if requestBody.OutputJSON != nil {
		outputJSON = *requestBody.OutputJSON
	}

	// Crear el cuerpo de la solicitud para OpenAI
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

	// Hacer la solicitud a OpenAI
	response, statusCode, err := OpenAIResponseJSON(oaiRequestBody)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Devolver la respuesta al cliente
	return c.Status(statusCode).Send(response)
}

