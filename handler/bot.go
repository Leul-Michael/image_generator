package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Leul-Michael/image-generation/model"
	repository "github.com/Leul-Michael/image-generation/repository/user"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

type BotHandler struct {
	bot *telebot.Bot
	db  *gorm.DB
}

// User states for different flows
type UserState struct {
	State      string
	CategoryID string
	PromptText string
}

var userStates = make(map[int64]*UserState)

func NewBotHandler(bot *telebot.Bot, db *gorm.DB) *BotHandler {
	return &BotHandler{
		bot: bot,
		db:  db,
	}
}

func (h *BotHandler) RegisterHandlers() {
	h.bot.Handle("/start", h.handleStart)
	h.bot.Handle("/cancel", h.handleCancel)
	h.bot.Handle("generate_image", h.handleGenerateImage)
	h.bot.Handle("my_credits", h.handleMyCredits)
	h.bot.Handle("trending_prompts", h.handleTrendingPrompts)
	h.bot.Handle("help", h.handleHelp)
	h.bot.Handle("back_to_main", h.handleBackToMain)
	h.bot.Handle("deposit_credits", h.handleDepositCredits)

	// Handle text messages for various inputs
	h.bot.Handle(telebot.OnText, h.handleTextMessage)

	// Handle all callback queries
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

	// Check if this is from a callback (has a callback query)
	if c.Callback() != nil {
		return c.Edit(welcomeMsg, menu)
	}

	// Otherwise, send a new message
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
	var imageCredits int
	for _, credit := range user.UserCredits {
		switch credit.CreditType {
		case model.CreditTypeImage:
			imageCredits = credit.Credits
		}
	}

	menu := &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{Text: "💰 Deposit Credits", Data: "deposit_credits"},
			},
			{
				{Text: "🔙 Back to Main Menu", Data: "back_to_main"},
			},
		},
	}

	message := fmt.Sprintf(
		"💳 Your Credit Balance:\n\n"+
			"🎨 Image Credits: %d\n\n"+
			"💡 Credit Pricing:\n"+
			"• 10 etb = 1 Image Credit\n"+
			"• 20 etb = 2 Image Credits\n"+
			"• 30 etb = 3 Image Credits\n"+
			"• And so on...\n\n"+
			"Credits are used to generate amazing AI images!",
		imageCredits,
	)

	if c.Callback() != nil {
		return c.Edit(message, menu)
	}
	return c.Send(message, menu)
}

func (h *BotHandler) handleDepositCredits(c telebot.Context) error {
	sender := c.Sender()
	if sender == nil {
		return fmt.Errorf("failed to get sender information")
	}

	// Set user state to waiting for deposit amount
	userStates[sender.ID] = &UserState{State: "waiting_deposit_amount"}

	message := "💰 Deposit Credits\n\n" +
		"Please enter the amount you want to deposit:\n\n" +
		"💡 Credit Conversion:\n" +
		"• 10 etb = 1 Image Credit\n" +
		"• 20 etb = 2 Image Credits\n" +
		"• 30 etb = 3 Image Credits\n\n" +
		"⚠️ Note: Only multiples of 10 are converted to credits.\n" +
		"For example: If you deposit 15, only 10 will be used (1 credit).\n\n" +
		"💬 Type your deposit amount or use /cancel to cancel:"

	if c.Callback() != nil {
		return c.Edit(message)
	}
	return c.Send(message)
}

func (h *BotHandler) handleTrendingPrompts(c telebot.Context) error {
	// Get trending prompts from database
	var trendingPrompts []model.TrendingPrompt
	if err := h.db.Where("is_active = ?", true).
		Preload("Category").
		Order("use_count DESC").
		Limit(10).
		Find(&trendingPrompts).Error; err != nil {

		return c.Send("❌ Could not load trending prompts. Please try again later.")
	}

	if len(trendingPrompts) == 0 {
		// No trending prompts available
		menu := &telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{
				{
					{Text: "🔙 Back to Main Menu", Data: "back_to_main"},
				},
			},
		}

		message := "📊 Trending Prompts\n\n" +
			"🤷‍♂️ Nothing trending at the moment.\n\n" +
			"Be the first to create some amazing images and start new trends!"

		if c.Callback() != nil {
			return c.Edit(message, menu)
		}
		return c.Send(message, menu)
	}

	var rows [][]telebot.InlineButton
	for _, prompt := range trendingPrompts {
		displayText := prompt.Prompt
		emojiText := prompt.Category.Emoji
		if len(displayText) > 35 {
			displayText = displayText[:32] + "..."
		}

		rows = append(rows, []telebot.InlineButton{
			{
				Text: fmt.Sprintf("%s %s", emojiText, displayText),
				Data: fmt.Sprintf("trending_%s", prompt.ID.String()),
			},
		})
	}

	// Add back button
	rows = append(rows, []telebot.InlineButton{
		{Text: "🔙 Back to Main Menu", Data: "back_to_main"},
	})

	menu := &telebot.ReplyMarkup{
		InlineKeyboard: rows,
	}

	message := "📊 Trending Prompts\n\n" +
		"Choose a popular prompt to use for image generation:\n\n"

	if c.Callback() != nil {
		return c.Edit(message, menu)
	}
	return c.Send(message, menu)
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
	case "deposit_credits":
		return h.handleDepositCredits(c)
	}

	// Handle category selection
	if len(data) > 9 && data[:9] == "category_" {
		categoryID := data[9:]
		return h.handleCategorySelected(c, categoryID)
	}

	// Handle trending prompt selection
	if len(data) > 9 && data[:9] == "trending_" {
		promptID := data[9:]
		return h.handleTrendingPromptSelected(c, promptID)
	}

	return nil
}

func (h *BotHandler) handleCategorySelected(c telebot.Context, categoryID string) error {
	// Get the category details
	var category model.Category
	if err := h.db.Where("id = ?", categoryID).First(&category).Error; err != nil {
		return c.Send("❌ Invalid category selected. Please try again.")
	}

	sender := c.Sender()
	if sender == nil {
		return fmt.Errorf("failed to get sender information")
	}

	// Set user state to waiting for prompt input
	userStates[sender.ID] = &UserState{
		State:      "waiting_prompt",
		CategoryID: categoryID,
	}

	emoji := "🎨"
	if category.Emoji != "" {
		emoji = category.Emoji
	}

	message := fmt.Sprintf(
		"%s %s Category Selected!\n\n"+
			"%s\n\n"+
			"✍️ Now, please describe the image you want to generate:\n\n"+
			"💡 Examples:\n"+
			"• A cute golden retriever puppy playing in a sunny meadow\n"+
			"• A futuristic city with flying cars at night\n"+
			"• A cozy cottage by a lake with mountains in the background\n\n"+
			"💬 Type your prompt or use /cancel to go back:",
		emoji,
		category.Name,
		category.Description,
	)

	return c.Edit(message)
}

func (h *BotHandler) handleTrendingPromptSelected(c telebot.Context, promptID string) error {
	sender := c.Sender()
	if sender == nil {
		return fmt.Errorf("failed to get sender information")
	}

	// Get from database
	var trendingPrompt model.TrendingPrompt
	if err := h.db.Where("id = ?", promptID).First(&trendingPrompt).Error; err != nil {
		return c.Send("❌ Invalid prompt selected. Please try again.")
	}

	// Update use count
	trendingPrompt.UseCount++
	trendingPrompt.LastUsedAt = time.Now()
	h.db.Save(&trendingPrompt)

	// Generate image with the selected prompt
	return h.generateImageWithPrompt(c, trendingPrompt.Prompt, "trending")
}

func (h *BotHandler) handleTextMessage(c telebot.Context) error {
	sender := c.Sender()
	if sender == nil {
		return nil
	}

	// Check if user is in any flow
	state, exists := userStates[sender.ID]
	if !exists {
		return nil // Ignore text messages if not in any flow
	}

	text := strings.TrimSpace(c.Text())

	switch state.State {
	case "waiting_deposit_amount":
		return h.handleDepositAmountInput(c, text)
	case "waiting_prompt":
		return h.handlePromptInput(c, text, state.CategoryID)
	}

	return nil
}

func (h *BotHandler) handleDepositAmountInput(c telebot.Context, text string) error {
	sender := c.Sender()

	// Validate input
	amount, err := strconv.Atoi(text)
	if err != nil {
		return c.Send("❌ Invalid input! Please enter a valid number.\n\n💬 Try again or use /cancel:")
	}

	if amount <= 0 {
		return c.Send("❌ Amount must be positive! Please enter a positive number.\n\n💬 Try again or use /cancel:")
	}

	if amount < 10 {
		return c.Send("❌ Minimum deposit is 10 etb to get 1 credit!\n\n💬 Please enter at least 10 or use /cancel:")
	}

	// Calculate credits (only multiples of 10)
	creditsToAdd := amount / 10
	unusedAmount := amount % 10

	// Clear user state
	delete(userStates, sender.ID)

	// Process the deposit
	return h.processDeposit(c, amount, creditsToAdd, unusedAmount)
}

func (h *BotHandler) handlePromptInput(c telebot.Context, text string, categoryID string) error {
	sender := c.Sender()

	if len(text) < 5 {
		return c.Send("❌ Please provide a more detailed description (at least 5 characters).\n\n💬 Try again or use /cancel:")
	}

	if len(text) > 500 {
		return c.Send("❌ Description is too long! Please keep it under 500 characters.\n\n💬 Try again or use /cancel:")
	}

	// Clear user state
	delete(userStates, sender.ID)

	// Generate image with the prompt
	return h.generateImageWithPrompt(c, text, categoryID)
}

func (h *BotHandler) generateImageWithPrompt(c telebot.Context, prompt string, categoryID string) error {
	sender := c.Sender()
	if sender == nil {
		return fmt.Errorf("failed to get sender information")
	}

	// Simulate processing time
	time.Sleep(2 * time.Second)

	// Get category name for display
	var categoryName string
	if categoryID == "trending" {
		categoryName = "Trending"
	} else {
		var category model.Category
		if err := h.db.Where("id = ?", categoryID).First(&category).Error; err == nil {
			categoryName = category.Name
		} else {
			categoryName = "Unknown"
		}
	}

	// Create placeholder response (since AI is not integrated yet)
	menu := &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{Text: "🔄 Generate Another", Data: "generate_image"},
				{Text: "📊 Use Trending", Data: "trending_prompts"},
			},
			{
				{Text: "💳 My Credits", Data: "my_credits"},
				{Text: "🏠 Main Menu", Data: "back_to_main"},
			},
		},
	}

	// Placeholder response
	message := fmt.Sprintf(
		"✅ Image Generated Successfully!\n\n"+
			"📝 Prompt: %s\n"+
			"📂 Category: %s\n"+
			"🎨 Style: AI Generated\n"+
			"⏱️ Generation Time: 2.3 seconds\n\n"+
			"🖼️ [Placeholder: Your amazing AI-generated image would appear here]\n\n"+
			"💡 This is a placeholder response. Once AI integration is complete, you'll see your actual generated image here!\n\n"+
			"What would you like to do next?",
		prompt,
		categoryName,
	)

	return c.Edit(message, menu)
}

func (h *BotHandler) handleCancel(c telebot.Context) error {
	sender := c.Sender()
	if sender == nil {
		return fmt.Errorf("failed to get sender information")
	}

	// Clear user state
	delete(userStates, sender.ID)

	return c.Send("❌ Operation cancelled.\n\nReturning to main menu...", &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{Text: "🏠 Main Menu", Data: "back_to_main"},
			},
		},
	})
}

func (h *BotHandler) processDeposit(c telebot.Context, amount, creditsToAdd, unusedAmount int) error {
	sender := c.Sender()
	if sender == nil {
		return fmt.Errorf("failed to get sender information")
	}

	userRepo := &repository.PostgresUserRepo{DB: h.db}
	user, err := userRepo.GetByTelegramID(context.TODO(), uint(sender.ID))
	if err != nil {
		return c.Send("❌ Could not process your deposit. Please try again.")
	}

	// Debug logging
	fmt.Printf("Processing deposit for user ID: %s, amount: %d, credits: %d\n", user.ID, amount, creditsToAdd)

	// Update user credits in database with proper transaction handling
	tx := h.db.Begin()
	if tx.Error != nil {
		fmt.Printf("Failed to begin transaction: %v\n", tx.Error)
		return c.Send("❌ Database error. Please try again.")
	}

	// Find existing user credit record
	var userCredit model.UserCredit
	result := tx.Where("user_id = ? AND credit_type = ?", user.ID, model.CreditTypeImage).First(&userCredit)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new credit record
			fmt.Printf("Creating new credit record for user: %s\n", user.ID)
			userCredit = model.UserCredit{
				UserID:     user.ID,
				CreditType: model.CreditTypeImage,
				Credits:    creditsToAdd,
			}
			if err := tx.Create(&userCredit).Error; err != nil {
				tx.Rollback()
				fmt.Printf("Failed to create credit record: %v\n", err)
				return c.Send("❌ Failed to process deposit. Please try again.")
			}
			fmt.Printf("Created new credit record with ID: %s\n", userCredit.ID)
		} else {
			tx.Rollback()
			fmt.Printf("Database error finding credit record: %v\n", result.Error)
			return c.Send("❌ Database error. Please try again.")
		}
	} else {
		// Update existing credit record
		fmt.Printf("Updating existing credit record. Current credits: %d, adding: %d\n", userCredit.Credits, creditsToAdd)
		userCredit.Credits += creditsToAdd
		if err := tx.Save(&userCredit).Error; err != nil {
			tx.Rollback()
			fmt.Printf("Failed to update credit record: %v\n", err)
			return c.Send("❌ Failed to process deposit. Please try again.")
		}
		fmt.Printf("Updated credit record. New balance: %d\n", userCredit.Credits)
	}

	// Create transaction record
	transaction := model.Transaction{
		UserID:       user.ID,
		CreditType:   model.CreditTypeImage,
		Amount:       creditsToAdd,
		Type:         model.TransactionTypePurchase,
		Description:  fmt.Sprintf("Deposit: %d etb converted to %d credits", amount-unusedAmount, creditsToAdd),
		BalanceAfter: userCredit.Credits,
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		fmt.Printf("Failed to create transaction record: %v\n", err)
		return c.Send("❌ Failed to record transaction. Please try again.")
	}
	fmt.Printf("Created transaction record with ID: %s\n", transaction.ID)

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		fmt.Printf("Failed to commit transaction: %v\n", err)
		return c.Send("❌ Failed to complete deposit. Please try again.")
	}

	fmt.Printf("Deposit completed successfully. Final balance: %d\n", userCredit.Credits)

	// Create success message
	menu := &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{Text: "💳 View Credits", Data: "my_credits"},
				{Text: "🏠 Main Menu", Data: "back_to_main"},
			},
		},
	}

	var message string
	if unusedAmount > 0 {
		message = fmt.Sprintf(
			"✅ Deposit Successful!\n\n"+
				"💰 Amount Deposited: %d etb\n"+
				"🎨 Credits Added: %d\n"+
				"💔 Unused Amount: %d etb\n\n"+
				"💳 Your New Balance: %d credits\n\n"+
				"⚠️ Note: %d etb were not converted because you need multiples of 10 for credits.",
			amount, creditsToAdd, unusedAmount, userCredit.Credits, unusedAmount,
		)
	} else {
		message = fmt.Sprintf(
			"✅ Deposit Successful!\n\n"+
				"💰 Amount Deposited: %d etb\n"+
				"🎨 Credits Added: %d\n\n"+
				"💳 Your New Balance: %d credits",
			amount, creditsToAdd, userCredit.Credits,
		)
	}

	return c.Send(message, menu)
}
