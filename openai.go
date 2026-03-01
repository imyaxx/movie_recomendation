package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Message represents a single turn in a chat conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest is the JSON body sent to the OpenAI API.
type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// chatResponse is the JSON body returned by the OpenAI API.
type chatResponse struct {
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// Client sends requests to the OpenAI Chat Completions API.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates an API client using the provided key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: APITimeout},
	}
}

// Send sends the full conversation history to the API and returns the assistant's reply.
func (c *Client) Send(messages []Message) (string, error) {
	body := chatRequest{
		Model:       DefaultModel,
		Messages:    messages,
		Temperature: DefaultTemperature,
		MaxTokens:   DefaultMaxTokens,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, APIEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("OpenAI error [%s]: %s", result.Error.Type, result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("API returned no choices")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

