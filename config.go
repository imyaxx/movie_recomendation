package main

import "time"

const (
	// --- API ---

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

	// DefaultMaxQuestions is the maximum number of interview questions before nudging the AI.
	DefaultMaxQuestions = 5

	// RecommendationMarker is the sentinel string the AI must begin its final response with.
	RecommendationMarker = "RECOMMENDATIONS:"

	// --- UI ---

	// botIcon is the emoji shown next to the bot name throughout the interface.
	botIcon = "🎬"

	// divider is the horizontal rule printed after the welcome banner.
	divider = "  ─────────────────────────────────────────────────────"

	// inputPrompt is the cursor displayed when prompting the user to choose an option.
	inputPrompt = "  › "

	// langChoicePrompt is shown before the language selection menu (language-neutral by design).
	langChoicePrompt = "  Choose language / Выберите язык:"

	// langOptionEN and langOptionRU are the two options in the language selection menu.
	langOptionEN = "  1 · English"
	langOptionRU = "  2 · Русский"
)
