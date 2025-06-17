package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	repository "github.com/Leul-Michael/image-generation/repository/user"
	"github.com/gin-gonic/gin"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
}

type TelegramInitData struct {
	QueryID  string       `json:"query_id"`
	User     TelegramUser `json:"user"`
	AuthDate int64        `json:"auth_date"`
	Hash     string       `json:"hash"`
}

type UserHandler struct {
	repo *repository.PostgresUserRepo
	bot  *telebot.Bot
}

func NewUserHandler(db *gorm.DB, bot *telebot.Bot) *UserHandler {
	return &UserHandler{
		repo: &repository.PostgresUserRepo{DB: db},
		bot:  bot,
	}
}

func (h *UserHandler) VerifyTelegramInitData(initData string) (bool, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return false, fmt.Errorf("failed to parse init data: %v", err)
	}

	hash := values.Get("hash")
	values.Del("hash")

	var keys []string
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var dataCheckString strings.Builder
	for i, k := range keys {
		if i > 0 {
			dataCheckString.WriteString("\n")
		}
		dataCheckString.WriteString(k)
		dataCheckString.WriteString("=")
		dataCheckString.WriteString(values.Get(k))
	}

	secretKey := sha256.Sum256([]byte("WebAppData"))
	hmacHash := hmac.New(sha256.New, secretKey[:])
	hmacHash.Write([]byte(dataCheckString.String()))
	calculatedHash := hex.EncodeToString(hmacHash.Sum(nil))

	return calculatedHash == hash, nil
}

func (h *UserHandler) HandleTelegramAuth(c *gin.Context) {
	initData := c.Query("initData")
	if initData == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "initData is required"})
		return
	}

	valid, err := h.VerifyTelegramInitData(initData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("verification failed: %v", err)})
		return
	}
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid init data"})
		return
	}

	values, err := url.ParseQuery(initData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to parse init data: %v", err)})
		return
	}

	userData := values.Get("user")
	var telegramUser TelegramUser
	if err := json.Unmarshal([]byte(userData), &telegramUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to parse user data: %v", err)})
		return
	}

	chat, err := h.bot.ChatByID(telegramUser.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Telegram user"})
		return
	}

	// Create or update user using repository
	user, err := h.repo.CreateOrUpdateUser(
		c.Request.Context(),
		uint(telegramUser.ID),
		telegramUser.FirstName,
		telegramUser.LastName,
		&telegramUser.Username,
		&telegramUser.PhotoURL,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create/update user: %v", err)})
		return
	}

	// Send welcome message for new users
	if user.CreatedAt.Equal(user.UpdatedAt) {
		_, err = h.bot.Send(chat, "Welcome to Image Generation AI! Your account has been created successfully.")
		if err != nil {
			fmt.Printf("Failed to send welcome message: %v\n", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Authentication successful",
		"user":    user,
	})
}
