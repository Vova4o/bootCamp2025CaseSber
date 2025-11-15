package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// API request/response structures
type SearchRequest struct {
	Query     string `json:"query"`
	Mode      string `json:"mode"`
	SessionID string `json:"session_id,omitempty"`
}

type SearchResponse struct {
	Answer    string   `json:"answer"`
	Sources   []Source `json:"sources"`
	SessionID string   `json:"session_id,omitempty"`
	Mode      string   `json:"mode,omitempty"`
}

type Source struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Snippet string  `json:"snippet"`
	Score   float64 `json:"score,omitempty"`
}

// User session management
type UserSession struct {
	SessionID string
	Mode      string
}

var userSessions = make(map[int64]*UserSession)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	log.Printf("🔑 Bot token: %s", botToken) // Debug: show what token we got

	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}
	log.Println(botToken)

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8000"
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("✅ Bot authorized as @%s", bot.Self.UserName)
	log.Printf("🔗 Using API: %s", apiURL)

	// Set up menu buttons
	setupMenuButtons(bot)

	// Delete webhook if set (use long polling instead)
	deleteWebhook := tgbotapi.DeleteWebhookConfig{DropPendingUpdates: true}
	_, err = bot.Request(deleteWebhook)
	if err != nil {
		log.Printf("Warning: Failed to delete webhook: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	log.Printf("📡 Starting to listen for updates...")
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		log.Printf("📥 Received update: %+v", update.UpdateID)

		if update.Message != nil {
			chatID := update.Message.Chat.ID
			userID := update.Message.From.ID
			text := update.Message.Text

			log.Printf("💬 Message from user %d: %s", userID, text)

			// Handle commands
			if update.Message.IsCommand() {
				handleCommand(bot, update.Message, userID)
				continue
			}

			// Handle button presses (reply keyboard)
			switch text {
			case "🔧 Выбрать режим":
				handleModeButton(bot, chatID, userID)
				continue
			case "🆕 Новая сессия":
				handleNewSessionButton(bot, chatID, userID)
				continue
			case "❓ Помощь":
				handleHelpButton(bot, chatID)
				continue
			}

			// Handle regular messages (search queries)
			go handleQuery(bot, chatID, userID, text, apiURL)
		}

		// Handle callback queries (button clicks)
		if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery)
		}
	}
}

func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, userID int64) {
	chatID := msg.Chat.ID

	switch msg.Command() {
	case "start":
		welcomeText := `👋 Привет! Я бот для интеллектуального поиска.

🔍 *Режимы работы:*
• *Simple* - быстрый поиск фактов
• *Pro* - глубокий анализ с контекстом
• *Auto* - автоматический выбор режима

Просто отправь мне вопрос, и я найду ответ! 🚀

Используй кнопки внизу для управления ботом 👇`

		// Create reply keyboard with buttons at the bottom
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("🔧 Выбрать режим"),
				tgbotapi.NewKeyboardButton("🆕 Новая сессия"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("❓ Помощь"),
			),
		)
		keyboard.ResizeKeyboard = true

		reply := tgbotapi.NewMessage(chatID, welcomeText)
		reply.ParseMode = "Markdown"
		reply.ReplyMarkup = keyboard
		bot.Send(reply)

		// Initialize session
		if userSessions[userID] == nil {
			userSessions[userID] = &UserSession{
				SessionID: "",
				Mode:      "auto",
			}
		}

	case "mode":
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🤖 Auto", "mode_auto"),
				tgbotapi.NewInlineKeyboardButtonData("⚡ Simple", "mode_simple"),
				tgbotapi.NewInlineKeyboardButtonData("🚀 Pro", "mode_pro"),
			),
		)

		currentMode := "auto"
		if session, ok := userSessions[userID]; ok {
			currentMode = session.Mode
		}

		text := fmt.Sprintf("Текущий режим: *%s*\n\nВыберите новый режим:", currentMode)
		reply := tgbotapi.NewMessage(chatID, text)
		reply.ParseMode = "Markdown"
		reply.ReplyMarkup = keyboard
		bot.Send(reply)

	case "newsession":
		if session, ok := userSessions[userID]; ok {
			session.SessionID = ""
		}

		// Send confirmation with keyboard
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("🔧 Выбрать режим"),
				tgbotapi.NewKeyboardButton("🆕 Новая сессия"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("❓ Помощь"),
			),
		)
		keyboard.ResizeKeyboard = true

		reply := tgbotapi.NewMessage(chatID, "✅ Новая сессия начата. История разговора очищена.")
		reply.ReplyMarkup = keyboard
		bot.Send(reply)

	case "help":
		helpText := `❓ *Помощь*

*Как использовать:*
1. Нажми кнопку "🔧 Выбрать режим" внизу
2. Отправь свой вопрос текстом
3. Получи ответ с источниками

*Режимы:*
• *Auto* - бот сам выберет лучший режим
• *Simple* - для простых вопросов (Кто? Что? Когда?)
• *Pro* - для сложных вопросов с контекстом беседы

*Примеры вопросов:*
• "Кто изобрел телефон?"
• "Сравни экономики США и Китая"
• "Как изменение климата влияет на сельское хозяйство?"

📞 *Обратная связь:* @yourusername`

		// Send with keyboard
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("🔧 Выбрать режим"),
				tgbotapi.NewKeyboardButton("🆕 Новая сессия"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("❓ Помощь"),
			),
		)
		keyboard.ResizeKeyboard = true

		reply := tgbotapi.NewMessage(chatID, helpText)
		reply.ParseMode = "Markdown"
		reply.ReplyMarkup = keyboard
		bot.Send(reply)

	default:
		reply := tgbotapi.NewMessage(chatID, "❌ Неизвестная команда. Используй /help")
		bot.Send(reply)
	}
}

func handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Parse callback data
	data := callback.Data
	if strings.HasPrefix(data, "mode_") {
		mode := strings.TrimPrefix(data, "mode_")

		// Update user session
		session, ok := userSessions[userID]
		if !ok {
			session = &UserSession{SessionID: "", Mode: mode}
			userSessions[userID] = session
		} else {
			session.Mode = mode
			session.SessionID = "" // Reset session when changing mode
		}

		// Send confirmation
		text := fmt.Sprintf("✅ Режим изменен на: *%s*", mode)
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		bot.Send(msg)

		// Answer callback to remove loading state
		bot.Request(tgbotapi.NewCallback(callback.ID, "Режим изменен"))
	}
}

func handleQuery(bot *tgbotapi.BotAPI, chatID int64, userID int64, query string, apiURL string) {
	log.Printf("🔍 Processing query: %s", query)

	// Show typing indicator
	typingAction := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	bot.Send(typingAction)

	// Get or create user session
	session, ok := userSessions[userID]
	if !ok {
		session = &UserSession{
			SessionID: "",
			Mode:      "auto",
		}
		userSessions[userID] = session
	}

	// Create backend session if we don't have one
	if session.SessionID == "" {
		sessionID, err := createChatSession(apiURL, session.Mode)
		if err != nil {
			log.Printf("❌ Failed to create session: %v", err)
			errorMsg := tgbotapi.NewMessage(chatID, "❌ Ошибка создания сессии")
			bot.Send(errorMsg)
			return
		}
		session.SessionID = sessionID
		log.Printf("✅ Created new chat session: %s", sessionID)
	}

	log.Printf("📤 Calling API with session: %s, mode: %s", session.SessionID, session.Mode)

	// Call chat session endpoint (maintains context)
	response, err := sendChatMessage(apiURL, session.SessionID, query, session.Mode)
	if err != nil {
		log.Printf("❌ API Error: %v", err)
		errorMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Ошибка: %v", err))
		bot.Send(errorMsg)
		return
	}

	log.Printf("✅ Got response: %d sources", len(response.Sources))

	// Format and send response
	responseText := formatResponse(response)
	log.Printf("📝 Formatted response (%d chars)", len(responseText))

	msg := tgbotapi.NewMessage(chatID, responseText)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	sentMsg, err := bot.Send(msg)
	if err != nil {
		log.Printf("❌ Failed to send message: %v", err)
		// Try without markdown
		msg.ParseMode = ""
		msg.Text = fmt.Sprintf("💬 Ответ:\n%s", response.Answer)
		bot.Send(msg)
	} else {
		log.Printf("✅ Message sent successfully: %d", sentMsg.MessageID)
	}
}

// Create a new chat session
func createChatSession(apiURL, mode string) (string, error) {
	reqBody := map[string]string{"mode": mode}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(apiURL+"/api/chat/session", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create session: status %d", resp.StatusCode)
	}

	var sessionResp struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return "", err
	}

	return sessionResp.ID, nil
}

// Send message to existing chat session
func sendChatMessage(apiURL, sessionID, query, mode string) (*SearchResponse, error) {
	reqBody := map[string]string{
		"query": query,
		"mode":  mode,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/chat/session/%s/message", apiURL, sessionID)
	log.Printf("🌐 POST %s", url)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("📡 API Response Status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	return &searchResp, nil
}

func formatResponse(resp *SearchResponse) string {
	var builder strings.Builder

	// Answer (no markdown escaping for regular text with Cyrillic)
	builder.WriteString("💬 *Ответ:*\n")
	builder.WriteString(resp.Answer)
	builder.WriteString("\n\n")

	// Sources
	if len(resp.Sources) > 0 {
		builder.WriteString("📚 *Источники:*\n")
		for i, source := range resp.Sources {
			if i >= 5 { // Limit to 5 sources to avoid message length issues
				builder.WriteString(fmt.Sprintf("\n...и ещё %d источников", len(resp.Sources)-i))
				break
			}
			builder.WriteString(fmt.Sprintf("%d. %s\n%s\n\n",
				i+1,
				truncate(source.Title, 80),
				source.URL))
		}
	}

	// Mode indicator
	if resp.Mode != "" {
		builder.WriteString(fmt.Sprintf("\n🔧 Режим: *%s*", resp.Mode))
	}

	return builder.String()
}

func escapeMarkdown(text string) string {
	// Escape special Markdown characters for Telegram MarkdownV2
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func handleModeButton(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🤖 Auto", "mode_auto"),
			tgbotapi.NewInlineKeyboardButtonData("⚡ Simple", "mode_simple"),
			tgbotapi.NewInlineKeyboardButtonData("🚀 Pro", "mode_pro"),
		),
	)

	currentMode := "auto"
	if session, ok := userSessions[userID]; ok {
		currentMode = session.Mode
	}

	text := fmt.Sprintf("Текущий режим: *%s*\n\nВыберите новый режим:", currentMode)
	reply := tgbotapi.NewMessage(chatID, text)
	reply.ParseMode = "Markdown"
	reply.ReplyMarkup = keyboard
	bot.Send(reply)
}

func handleNewSessionButton(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
	// Clear the session ID so a new one will be created on next message
	if session, ok := userSessions[userID]; ok {
		session.SessionID = ""
	}

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔧 Выбрать режим"),
			tgbotapi.NewKeyboardButton("🆕 Новая сессия"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❓ Помощь"),
		),
	)
	keyboard.ResizeKeyboard = true

	reply := tgbotapi.NewMessage(chatID, "✅ Новая сессия начата. История разговора очищена.")
	reply.ReplyMarkup = keyboard
	bot.Send(reply)
}

func handleHelpButton(bot *tgbotapi.BotAPI, chatID int64) {
	helpText := `❓ *Помощь*

*Как использовать:*
1. Нажми кнопку "🔧 Выбрать режим" внизу
2. Отправь свой вопрос текстом
3. Получи ответ с источниками

*Режимы:*
• *Auto* - бот сам выберет лучший режим
• *Simple* - для простых вопросов (Кто? Что? Когда?)
• *Pro* - для сложных вопросов с контекстом беседы

*Примеры вопросов:*
• "Кто изобрел телефон?"
• "Сравни экономики США и Китая"
• "Как изменение климата влияет на сельское хозяйство?"

📞 *Обратная связь:* @yourusername`

	// Send with keyboard
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔧 Выбрать режим"),
			tgbotapi.NewKeyboardButton("🆕 Новая сессия"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❓ Помощь"),
		),
	)
	keyboard.ResizeKeyboard = true

	reply := tgbotapi.NewMessage(chatID, helpText)
	reply.ParseMode = "Markdown"
	reply.ReplyMarkup = keyboard
	bot.Send(reply)
}

func setupMenuButtons(bot *tgbotapi.BotAPI) {
	// Create persistent menu buttons at the bottom of the chat
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "🏠 Начать работу"},
		{Command: "mode", Description: "🔧 Выбрать режим"},
		{Command: "newsession", Description: "🆕 Новая сессия"},
		{Command: "help", Description: "❓ Помощь"},
	}

	cfg := tgbotapi.NewSetMyCommands(commands...)
	bot.Request(cfg)
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		// File doesn't exist, skip loading
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		// Only set if not already set in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
