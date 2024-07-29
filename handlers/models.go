package handlers

import (
  "errors"
  "time"
)

type User struct {
  ID string `json:"id_user"`
  InputUsage float64 `json:"input_usage"`
  OutputUsage float64 `json:"output_usage"`
  History []History `json:"history"`
  OpenAIUsage Usage `json:"openai"`
  GoogleUsage Usage `json:"google"`
  AnthropicUsage Usage `json:"anthropic"`
}

type Usage struct {
  InputUsage float64 `json:"input_usage"`
  OutputUsage float64 `json:"output_usage"`
  Models map[string]ModelUsage `json:"models"`
}

type ModelUsage struct {
  InputTokens int `json:"input_tokens"`
  OutputTokens int `json:"output_tokens"`
  InputUsage float64 `json:"input_usage"`
  OutputUsage float64 `json:"output_usage"`
}

type History struct {
  Company string `json:"company"`
  Model string `json:"model"`
  InputTokens int `json:"input_tokens"`
  OutputTokens int `json:"output_tokens"`
  InputUsage float64 `json:"input_usage"`
  OutputUsage float64 `json:"output_usage"`
  Created time.Time `json:"created"`
}

// Structures and common functions

type RequestBody struct {
  ID string `json:"id_user"`
  Model string `json:"model"`
  SystemPrompt *string `json:"system_prompt,omitempty"`
  Prompt *string `json:"prompt,omitempty"`
  Messages []Message `json:"messages,omitempty"`
  OutputJSON *bool `json:"output_JSON"`
}

type Message struct {
  Role string `json:"role"`
  Content string `json:"content"`
}

func processRequest(body RequestBody) ([]OAIMessage, []Content, error) {
  var openAIMessages []OAIMessage
  var googleContents []Content

  if len(body.Messages) == 0 {
    systemPrompt := "You are a helpful assistant."
    if body.SystemPrompt != nil {
      systemPrompt = *body.SystemPrompt
    }
    
    if body.OutputJSON != nil && *body.OutputJSON {
      systemPrompt += "\nResponse Format: JSON"
    }

    if body.Prompt == nil {
      return nil, nil, errors.New("prompt is required if messages are not provided")
    }

    openAIMessages = append(openAIMessages, OAIMessage{Role: "system", Content: systemPrompt})
    openAIMessages = append(openAIMessages, OAIMessage{Role: "user", Content: *body.Prompt})

    googleContents = append(googleContents, Content{Role: "user", Parts: []Part{{Text: systemPrompt}}})
    googleContents = append(googleContents, Content{Role: "user", Parts: []Part{{Text: *body.Prompt}}})
  } else {
    if body.OutputJSON != nil && *body.OutputJSON {
      openAIMessages = append(openAIMessages, OAIMessage{Role: "system", Content: "Response Format: JSON"})
    }

    for _, msg := range body.Messages {
      openAIMessages = append(openAIMessages, OAIMessage{Role: msg.Role, Content: msg.Content})

      googleRole := "user"
      if msg.Role == "assistant" {
        googleRole = "model"
      }
      googleContents = append(googleContents, Content{Role: googleRole, Parts: []Part{{Text: msg.Content}}})
    }
  }

  return openAIMessages, googleContents, nil
}

// openai-specific structures
type OAIMessage struct {
  Role string `json:"role"`
  Content string `json:"content"`
}

type OAIRequestBody struct {
  Model string `json:"model"`
  Messages []OAIMessage `json:"messages"`
  ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

type ResponseFormat struct {
  Type string `json:"type"`
}

// google-specific structures
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
