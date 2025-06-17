package handler

import (
	"fmt"

	"gopkg.in/telebot.v3"
)

type BotHandler struct {
	bot *telebot.Bot
}

func NewBotHandler(bot *telebot.Bot) *BotHandler {
	fmt.Println("Bot handler created")
	return &BotHandler{
		bot: bot,
	}
}

func (h *BotHandler) RegisterHandlers() {
	// Start command handler
	fmt.Println("Registering handlers...")
	h.bot.Handle("/start", h.handleStart)
}

func (h *BotHandler) handleStart(c telebot.Context) error {
	fmt.Println("Start command received")
	// Create inline keyboard with buttons
	menu := &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{Text: "🎨 Generate Image", Data: "generate_image"},
				{Text: "💳 My Credits", Data: "my_credits"},
			},
			{
				{Text: "📊 Trending Prompts", Data: "trending_prompts"},
				{Text: "❓ Help", Data: "help"},
			},
		},
	}

	// Send welcome message with buttons
	return c.Send(
		fmt.Sprintf("Welcome to Image Generation AI Bot! 🎨\n\nI can help you generate amazing images using AI. What would you like to do?"),
		menu,
	)
}
