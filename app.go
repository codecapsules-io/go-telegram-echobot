package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	bot    *tgbotapi.BotAPI
	botErr error
	TOKEN  = os.Getenv("BOT_TOKEN")
	URL    = os.Getenv("URL")
)

func init() {
	log.Printf("URL: %s", URL)

	if TOKEN == "" {
		botErr = fmt.Errorf("BOT_TOKEN environment variable not set")
		log.Printf("Error: %v", botErr)
		return
	}

	var err error
	bot, err = tgbotapi.NewBotAPI(TOKEN)
	if err != nil {
		botErr = err
		log.Printf("Failed to create bot: %v", err)
		return
	}

	log.Printf("Bot initialized successfully")
}

func respond(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse the update from JSON
	var update tgbotapi.Update
	if err := json.Unmarshal(body, &update); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if update.Message == nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
		return
	}

	chatID := update.Message.Chat.ID
	messageID := update.Message.MessageID
	text := update.Message.Text

	// Handle /start command
	if text == "/start" {
		msg := tgbotapi.NewMessage(chatID, "Hi! I respond by echoing messages. Give it a try!")
		msg.ReplyToMessageID = messageID
		bot.Send(msg)
	} else {
		// Echo the message back
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyToMessageID = messageID
		bot.Send(msg)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

func setWebhook(w http.ResponseWriter, r *http.Request) {
	webhookURL := fmt.Sprintf("%s%s", URL, TOKEN)
	log.Printf("Setting webhook")

	// Direct HTTP call to Telegram API
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook?url=%s", TOKEN, webhookURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("Webhook error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "webhook setup failed")
		return
	}
	defer resp.Body.Close()

	log.Printf("Webhook response status: %d", resp.StatusCode)
	fmt.Fprint(w, "webhook setup ok")
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, welcome to the telegram bot index page")
}

func main() {
	// Register handlers
	http.HandleFunc(fmt.Sprintf("/%s", TOKEN), respond)
	http.HandleFunc("/setwebhook", setWebhook)
	http.HandleFunc("/", index)

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
