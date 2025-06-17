package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Leul-Michael/image-generation/handler"
	"github.com/Leul-Michael/image-generation/model"
	tele "gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

type App struct {
	router http.Handler
	DB     *gorm.DB
	bot    *tele.Bot
}

func New() (*App, error) {
	app := &App{}

	err := app.connectToDB()
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	app.DB.AutoMigrate(&model.User{}, &model.Category{}, &model.GeneratedImage{}, &model.ImageGenerationRequest{}, &model.Transaction{}, &model.UserCredit{}, &model.TrendingPrompt{})

	if err := app.SeedCategories(); err != nil {
		fmt.Printf("Warning: Failed to seed categories: %v\n", err)
	}

	err = app.connectToBot()
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	botHandler := handler.NewBotHandler(app.bot, app.DB)
	botHandler.RegisterHandlers()

	app.loadRoutes()

	return app, nil
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":5000",
		Handler: a.router,
	}

	ch := make(chan error, 1)

	go func() {
		fmt.Println("Starting bot...")
		a.bot.Start()
	}()

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := server.Shutdown(timeout); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
		a.bot.Stop()
	}
	return nil
}
