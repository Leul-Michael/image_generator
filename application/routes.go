package application

import (
	"github.com/gin-gonic/gin"
)

func (a *App) loadRoutes() {
	router := gin.Default()
	// api v1 route handler
	v1Router := router.Group("/api/v1")
	v1Router.Group("/auth")
	{

	}

	a.router = router
}
