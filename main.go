package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// botPrefix is the fixed label printed before every bot message (including "thinking").
// Keeping it as a constant ensures the cursor never shifts between states.
const botPrefix = "🎬 CineMatch"

// ui holds all user-visible strings for one language.
type ui struct {
	// banner lines (printed once at startup)
	bannerSub      string
	bannerLine1    string
	bannerLine2    string
	bannerLine3    string
	hint           string
	// conversation labels
	youLabel      string
	// status / feedback
	thinkingSuffix string // appended after botPrefix when waiting
	goodbye        string
	enjoy          string
	// errors
	errorConnect  string
	errorAPI      string
	// kickstart message sent silently to the API
	kickstart     string
	// language selection
	langInvalid   string
}

var textEN = ui{
	bannerSub:   "  Not sure what to watch tonight? Tell me how you feel\n  and I'll find the perfect film just for you.",
	bannerLine1: "  ✨  I'll ask you a few short questions:",
	bannerLine2: "  🎭  about your mood, favourite genres and recent films —",
	bannerLine3: "  🍿  and find something perfect just for you.",
	hint:        "  Type 'quit' at any time to exit.",
	youLabel:       "        You",
	thinkingSuffix: " is thinking...",
	goodbye:        "\n" + botPrefix + ": Have a great evening!\n",
	enjoy:          "\n  ✦ Enjoy the film!\n",
	errorConnect:   "Failed to connect: %v\n",
	errorAPI:       "API error: %v\n",
	kickstart:      "Hello, I'd like a movie recommendation.",
	langInvalid:    "  Please enter 1 or 2.",
}

var textRU = ui{
	bannerSub:   "  Не знаешь, что посмотреть сегодня вечером?\n  Расскажи мне о своём настроении — я подберу идеальный фильм.",
	bannerLine1: "  ✨  Я задам тебе несколько коротких вопросов:",
	bannerLine2: "  🎭  о настроении, любимых жанрах и последних фильмах —",
	bannerLine3: "  🍿  и подберу что-то идеальное именно для тебя.",
	hint:        "  Введи 'выйти' или 'quit' в любой момент, чтобы выйти.",
	youLabel:       "        Ты",
	thinkingSuffix: " думает...",
	goodbye:        "\n" + botPrefix + ": Хорошего вечера!\n",
	enjoy:          "\n  ✦ Приятного просмотра!\n",
	errorConnect:   "Не удалось подключиться: %v\n",
	errorAPI:       "Ошибка API: %v\n",
	kickstart:      "Привет, хочу получить рекомендацию фильма.",
	langInvalid:    "  Введите 1 или 2.",
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

	// Kick off the conversation — send the first message so the AI asks the opening question.
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

	// REPL: read user input, send to API, print response, repeat until recommendations arrive.
	for {
		fmt.Printf("\n%s: ", t.youLabel)

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

		// If the AI has been asking for a while, nudge it toward recommendations.
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

// chooseLang asks the user to pick a language and returns their choice.
func chooseLang(scanner *bufio.Scanner) Language {
	fmt.Println()
	fmt.Println("  Choose language / Выберите язык:")
	fmt.Println()
	fmt.Println("  1 · English")
	fmt.Println("  2 · Русский")
	fmt.Println()

	for {
		fmt.Print("  › ")
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

// printBanner prints the welcome header.
func printBanner(t ui) {
	fmt.Println()
	fmt.Println("  🎬  CineMatch")
	fmt.Println()
	fmt.Println(t.bannerSub)
	fmt.Println()
	fmt.Println(t.bannerLine1)
	fmt.Println(t.bannerLine2)
	fmt.Println(t.bannerLine3)
	fmt.Println()
	fmt.Println(t.hint)
	fmt.Println()
	printDivider()
}

// printDivider prints a subtle horizontal separator.
func printDivider() {
	fmt.Println("  ─────────────────────────────────────────────────────")
}

// printThinking prints a "thinking" line using the same fixed prefix as bot messages.
// This prevents any horizontal shift when the response arrives.
func printThinking(t ui) {
	fmt.Printf("\n%s%s", botPrefix, t.thinkingSuffix)
}

// clearThinking erases the thinking line so the real response can be printed in its place.
func clearThinking() {
	fmt.Print("\r\033[K")
}

// printBot prints the bot's response with the fixed prefix.
func printBot(message string) {
	fmt.Printf("\n%s: %s\n", botPrefix, message)
}

// isExitCommand reports whether the input is a quit command in any supported language.
func isExitCommand(input string) bool {
	lower := strings.ToLower(input)
	return lower == "quit" || lower == "exit" || lower == "q" ||
		lower == "выход" || lower == "выйти"
}

// loadEnvFile reads KEY=VALUE pairs from a .env file and sets them as environment variables.
// Lines starting with # and empty lines are ignored. Already-set variables are not overwritten.
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // .env is optional — no error if it doesn't exist
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
