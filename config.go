package main

import "time"

const (
	// EnvAPIKey is the environment variable name for the OpenAI API key.
	EnvAPIKey = "OPENAI_API_KEY"

	// APIEndpoint is the OpenAI Chat Completions endpoint.
	APIEndpoint = "https://api.openai.com/v1/chat/completions"

	// APITimeout is the maximum time to wait for an API response.
	APITimeout = 30 * time.Second

	// DefaultModel is the OpenAI model used for conversations.
	DefaultModel = "gpt-4o-mini"

	// DefaultTemperature controls response creativity (0.0 = deterministic, 1.0 = creative).
	DefaultTemperature = 0.8

	// DefaultMaxTokens limits the length of each API response.
	DefaultMaxTokens = 500

	// DefaultMaxQuestions is the maximum number of interview questions before forcing recommendations.
	DefaultMaxQuestions = 5

	// RecommendationMarker is the exact string the AI must start its recommendation response with.
	RecommendationMarker = "RECOMMENDATIONS:"

	// KickstartMessage is the initial user message that triggers the AI to ask its first question.
	KickstartMessage = "Hello, I'd like a movie recommendation."
)
