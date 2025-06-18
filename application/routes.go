package application

import (
	"net/http"

	"github.com/Leul-Michael/image-generation/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (a *App) loadRoutes() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "image-generation-ai, hulugram ai",
		})
	})

	userHandler := handler.NewUserHandler(a.DB, a.bot)

	v1Router := router.Group("/api/v1")
	{
		authRouter := v1Router.Group("/auth")
		{
			authRouter.GET("/telegram", userHandler.HandleTelegramAuth)
		}

		userRouter := v1Router.Group("/users")
		{
			userRouter.GET("/me", userHandler.GetCurrentUser)
			userRouter.PUT("/me", userHandler.UpdateCurrentUser)
			userRouter.GET("/me/credits", userHandler.GetUserCredits)
		}

		v1Router.GET("/categories", userHandler.GetCategories)
		v1Router.GET("/trending-prompts", userHandler.GetTrendingPrompts)
	}

	a.router = router
}
