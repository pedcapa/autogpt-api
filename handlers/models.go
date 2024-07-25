package handlers

import "time"

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
