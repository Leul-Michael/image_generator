package handler

import (
	"context"
	"fmt"

	"github.com/Leul-Michael/image-generation/model"
	repository "github.com/Leul-Michael/image-generation/repository/user"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

type BotHandler struct {
	bot *telebot.Bot
	db  *gorm.DB
}

func NewBotHandler(bot *telebot.Bot, db *gorm.DB) *BotHandler {
	return &BotHandler{
		bot: bot,
		db:  db,
	}
}

func (h *BotHandler) RegisterHandlers() {
	h.bot.Handle("/start", h.handleStart)
	h.bot.Handle(telebot.OnCallback, h.handleCallback)
}

func (h *BotHandler) handleStart(c telebot.Context) error {
	sender := c.Sender()

	if sender == nil {
		return fmt.Errorf("failed to get sender information")
	}

	userRepo := &repository.PostgresUserRepo{DB: h.db}
	user, err := userRepo.CreateOrUpdateUser(
		context.Background(),
		uint(sender.ID),
		sender.FirstName,
		sender.LastName,
		&sender.Username,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to process user data")
	}

	return h.sendMainMenu(c, user)
}

func (h *BotHandler) sendMainMenu(c telebot.Context, user *model.User) error {
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

	// Safely get credit balance
	var imageCredits int
	for _, credit := range user.UserCredits {
		if credit.CreditType == model.CreditTypeImage {
			imageCredits = credit.Credits
			break
		}
	}

	welcomeMsg := fmt.Sprintf(
		"Hello %s! Welcome to Image Generation Bot! 🎨\n\n"+
			"I can help you generate amazing images using AI. What would you like to do?\n\n"+
			"Your current image credits: %d",
		user.FirstName,
		imageCredits,
	)

	return c.Send(welcomeMsg, menu)
}

func (h *BotHandler) handleGenerateImage(c telebot.Context) error {
	// Get all active categories from database
	var categories []model.Category
	if err := h.db.Where("is_active = ?", true).Order("name ASC").Find(&categories).Error; err != nil {
		return c.Send("❌ Sorry, I couldn't load the categories. Please try again later.")
	}

	if len(categories) == 0 {
		return c.Send("❌ No categories available at the moment. Please try again later.")
	}

	// Create inline keyboard with categories (2 categories per row)
	var rows [][]telebot.InlineButton
	for i := 0; i < len(categories); i += 2 {
		var row []telebot.InlineButton

		// First category in the row
		emoji := "🎨"
		if categories[i].Emoji != "" {
			emoji = categories[i].Emoji
		}
		row = append(row, telebot.InlineButton{
			Text: fmt.Sprintf("%s %s", emoji, categories[i].Name),
			Data: fmt.Sprintf("category_%s", categories[i].ID.String()),
		})

		// Second category in the row (if exists)
		if i+1 < len(categories) {
			emoji2 := "🎨"
			if categories[i+1].Emoji != "" {
				emoji2 = categories[i+1].Emoji
			}
			row = append(row, telebot.InlineButton{
				Text: fmt.Sprintf("%s %s", emoji2, categories[i+1].Name),
				Data: fmt.Sprintf("category_%s", categories[i+1].ID.String()),
			})
		}

		rows = append(rows, row)
	}

	// Add back button
	rows = append(rows, []telebot.InlineButton{
		{Text: "🔙 Back to Main Menu", Data: "back_to_main"},
	})

	menu := &telebot.ReplyMarkup{
		InlineKeyboard: rows,
	}

	message := "🎨 Choose a category for your image generation:\n\n" +
		"Select the type of image you'd like to create!"

	return c.Edit(message, menu)
}

func (h *BotHandler) handleMyCredits(c telebot.Context) error {
	sender := c.Sender()
	if sender == nil {
		return fmt.Errorf("failed to get sender information")
	}

	userRepo := &repository.PostgresUserRepo{DB: h.db}
	user, err := userRepo.GetByTelegramID(context.TODO(), uint(sender.ID))
	if err != nil {
		return c.Send("❌ Could not retrieve your credit information.")
	}

	// Get credit balances
	var imageCredits, videoCredits int
	for _, credit := range user.UserCredits {
		switch credit.CreditType {
		case model.CreditTypeImage:
			imageCredits = credit.Credits
		case model.CreditTypeVideo:
			videoCredits = credit.Credits
		}
	}

	menu := &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{Text: "🔙 Back to Main Menu", Data: "back_to_main"},
			},
		},
	}

	message := fmt.Sprintf(
		"💳 Your Credit Balance:\n\n"+
			"🎨 Image Credits: %d\n"+
			"🎬 Video Credits: %d\n\n"+
			"Credits are used to generate images and videos. You can earn more credits through various activities!",
		imageCredits,
		videoCredits,
	)

	return c.Edit(message, menu)
}

func (h *BotHandler) handleTrendingPrompts(c telebot.Context) error {
	menu := &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{Text: "🔙 Back to Main Menu", Data: "back_to_main"},
			},
		},
	}

	message := "📊 Trending Prompts:\n\n" +
		"🔥 A magical forest with glowing mushrooms\n" +
		"🌟 Cyberpunk city at sunset\n" +
		"🏰 Medieval castle on a floating island\n" +
		"🦋 Butterfly garden in spring\n" +
		"🌊 Underwater palace with coral decorations\n\n" +
		"Coming soon: Dynamic trending prompts!"

	return c.Edit(message, menu)
}

func (h *BotHandler) handleHelp(c telebot.Context) error {
	menu := &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{Text: "🔙 Back to Main Menu", Data: "back_to_main"},
			},
		},
	}

	message := "❓ How to use the Image Generation Bot:\n\n" +
		"1️⃣ Click 'Generate Image' to start\n" +
		"2️⃣ Choose a category that interests you\n" +
		"3️⃣ Describe what you want to create\n" +
		"4️⃣ Wait for your AI-generated image!\n\n" +
		"💡 Tips:\n" +
		"• Be specific in your descriptions\n" +
		"• Use descriptive adjectives\n" +
		"• Mention colors, styles, or moods\n\n" +
		"Need help? Contact support!"

	return c.Edit(message, menu)
}

func (h *BotHandler) handleBackToMain(c telebot.Context) error {
	sender := c.Sender()
	if sender == nil {
		return fmt.Errorf("failed to get sender information")
	}

	userRepo := &repository.PostgresUserRepo{DB: h.db}
	user, err := userRepo.GetByTelegramID(context.TODO(), uint(sender.ID))
	if err != nil {
		return c.Send("❌ Could not retrieve your information.")
	}

	return h.sendMainMenu(c, user)
}

func (h *BotHandler) handleCallback(c telebot.Context) error {
	data := c.Callback().Data
	switch data {
	case "generate_image":
		return h.handleGenerateImage(c)
	case "my_credits":
		return h.handleMyCredits(c)
	case "trending_prompts":
		return h.handleTrendingPrompts(c)
	case "help":
		return h.handleHelp(c)
	case "back_to_main":
		return h.handleBackToMain(c)
	}
	if len(data) > 9 && data[:9] == "category_" {
		categoryID := data[9:]
		return h.handleCategorySelected(c, categoryID)
	}
	return nil
}

func (h *BotHandler) handleCategorySelected(c telebot.Context, categoryID string) error {
	// Get the category details
	var category model.Category
	if err := h.db.Where("id = ?", categoryID).First(&category).Error; err != nil {
		return c.Send("❌ Invalid category selected. Please try again.")
	}

	menu := &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{Text: "🔙 Back to Categories", Data: "generate_image"},
				{Text: "🏠 Main Menu", Data: "back_to_main"},
			},
		},
	}

	emoji := "🎨"
	if category.Emoji != "" {
		emoji = category.Emoji
	}

	message := fmt.Sprintf(
		"%s %s Category Selected!\n\n"+
			"%s\n\n"+
			"✍️ Now, please describe the image you want to generate in this category.\n\n"+
			"For example:\n"+
			"• Be creative and specific\n"+
			"• Mention colors, style, mood\n"+
			"• Describe the scene in detail\n\n"+
			"Type your prompt and I'll generate an amazing image for you!",
		emoji,
		category.Name,
		category.Description,
	)

	return c.Edit(message, menu)
}
