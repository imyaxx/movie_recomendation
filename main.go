package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// botPrefix is the label printed before every bot message.
// Defined as a constant so the cursor never shifts between "thinking" and reply states.
const botPrefix = botIcon + " CineMatch"

// ui holds all user-visible strings for a single language.
type ui struct {
	// Welcome banner
	bannerSub   string
	bannerLine1 string
	bannerLine2 string
	bannerLine3 string
	hint        string

	// Conversation
	youLabel       string
	thinkingSuffix string // appended to botPrefix while waiting for the API
	goodbye        string
	enjoy          string

	// Error messages (used with fmt.Fprintf + %v)
	errorConnect string
	errorAPI     string

	// Silent kickstart sent to the API to trigger the first question
	kickstart string
}

var textEN = ui{
	bannerSub:      "  Not sure what to watch tonight? Tell me how you feel\n  and I'll find the perfect film just for you.",
	bannerLine1:    "  ✨  I'll ask you a few short questions:",
	bannerLine2:    "  🎭  about your mood, favourite genres and recent films —",
	bannerLine3:    "  🍿  and find something perfect just for you.",
	hint:           "  Type 'quit' at any time to exit.",
	youLabel:       "You",
	thinkingSuffix: " is thinking...",
	goodbye:        "\n" + botPrefix + ": Have a great evening!\n",
	enjoy:          "\n  ✦ Enjoy the film!\n",
	errorConnect:   "Failed to connect: %v\n",
	errorAPI:       "API error: %v\n",
	kickstart:      "Hello, I'd like a movie recommendation.",
}

var textRU = ui{
	bannerSub:      "  Не знаешь, что посмотреть сегодня вечером?\n  Расскажи мне о своём настроении — я подберу идеальный фильм.",
	bannerLine1:    "  ✨  Я задам тебе несколько коротких вопросов:",
	bannerLine2:    "  🎭  о настроении, любимых жанрах и последних фильмах —",
	bannerLine3:    "  🍿  и подберу что-то идеальное именно для тебя.",
	hint:           "  Введи 'выйти' или 'quit' в любой момент, чтобы выйти.",
	youLabel:       "Ты",
	thinkingSuffix: " думает...",
	goodbye:        "\n" + botPrefix + ": Хорошего вечера!\n",
	enjoy:          "\n  ✦ Приятного просмотра!\n",
	errorConnect:   "Не удалось подключиться: %v\n",
	errorAPI:       "Ошибка API: %v\n",
	kickstart:      "Привет, хочу получить рекомендацию фильма.",
}

func main() {
	loadEnvFile(".env")

	apiKey := os.Getenv(EnvAPIKey)
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Error: %s is not set.\n", EnvAPIKey)
		fmt.Fprintf(os.Stderr, "Add it to .env or run: export %s=your-api-key\n", EnvAPIKey)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)

	lang := chooseLang(scanner)
	t := textEN
	if lang == LangRussian {
		t = textRU
	}

	client := NewClient(apiKey)
	session := NewSession(lang)

	printBanner(t)

	// Send the kickstart message silently so the AI opens with the first question.
	session.AddUserMessage(t.kickstart)
	printThinking(t)
	firstQuestion, err := client.Send(session.History)
	clearThinking()
	if err != nil {
		fmt.Fprintf(os.Stderr, t.errorConnect, err)
		os.Exit(1)
	}
	session.AddAssistantMessage(firstQuestion)
	printBot(firstQuestion)

	// Main loop: read input → send to API → print reply → repeat until recommendations arrive.
	for {
		fmt.Printf("\n        %s: ", t.youLabel)

		if !scanner.Scan() {
			fmt.Print(t.goodbye)
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if isExitCommand(input) {
			fmt.Print(t.goodbye)
			break
		}

		session.AddUserMessage(input)

		if session.ShouldNudge() {
			session.AddSystemNudge()
		}

		printThinking(t)
		response, err := client.Send(session.History)
		clearThinking()
		if err != nil {
			fmt.Fprintf(os.Stderr, t.errorAPI, err)
			break
		}
		session.AddAssistantMessage(response)
		printBot(response)

		if IsRecommendation(response) {
			fmt.Print(t.enjoy)
			break
		}
	}
}

// chooseLang shows a language selection menu and returns the user's choice.
// The menu text is language-neutral by design (stored in config.go).
func chooseLang(scanner *bufio.Scanner) Language {
	fmt.Println()
	fmt.Println(langChoicePrompt)
	fmt.Println()
	fmt.Println(langOptionEN)
	fmt.Println(langOptionRU)
	fmt.Println()

	for {
		fmt.Print(inputPrompt)
		if !scanner.Scan() {
			return LangEnglish
		}
		switch strings.TrimSpace(scanner.Text()) {
		case "1", "en", "english":
			return LangEnglish
		case "2", "ru", "русский":
			return LangRussian
		}
	}
}

// printBanner prints the welcome screen.
func printBanner(t ui) {
	fmt.Println()
	fmt.Printf("  %s  CineMatch\n", botIcon)
	fmt.Println()
	fmt.Println(t.bannerSub)
	fmt.Println()
	fmt.Println(t.bannerLine1)
	fmt.Println(t.bannerLine2)
	fmt.Println(t.bannerLine3)
	fmt.Println()
	fmt.Println(t.hint)
	fmt.Println()
	fmt.Println(divider)
}

// printBot prints the bot's response with the fixed prefix.
func printBot(message string) {
	fmt.Printf("\n%s: %s\n", botPrefix, message)
}

// printThinking shows a "thinking" indicator on the current line.
// Uses the same botPrefix as printBot so the cursor position never shifts.
func printThinking(t ui) {
	fmt.Printf("\n%s%s", botPrefix, t.thinkingSuffix)
}

// clearThinking erases the thinking indicator so the real response replaces it cleanly.
func clearThinking() {
	fmt.Print("\r\033[K")
}

// isExitCommand reports whether the input is a recognised quit command.
func isExitCommand(input string) bool {
	switch strings.ToLower(input) {
	case "quit", "exit", "q", "выход", "выйти":
		return true
	}
	return false
}

// loadEnvFile reads KEY=VALUE pairs from path and sets them as environment variables.
// Lines starting with # and blank lines are ignored.
// Variables already set in the environment are never overwritten.
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // .env is optional
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		os.Setenv(key, value)
	}
}
