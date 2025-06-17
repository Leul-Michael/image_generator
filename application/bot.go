package application

import (
	"fmt"
	"os"
	"time"

	tele "gopkg.in/telebot.v3"
)

func (a *App) connectToBot() error {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return fmt.Errorf("BOT_TOKEN environment variable is required")
	}

	// Initialize Telegram bot
	pref := tele.Settings{
		Token:  botToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		return fmt.Errorf("failed to create bot: %v", err)
	}

	a.bot = bot

	return nil
}
