package api

import (
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/api/handlers"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Initialize handlers
	searchHandler := handlers.NewSearchHandler(db, cfg)
	chatHandler := handlers.NewChatHandler(db, cfg)
	healthHandler := handlers.NewHealthHandler()

	// API routes
	api := router.Group("/api")
	{
		// Health check
		api.GET("/health", healthHandler.Health)

		// Search
		api.POST("/search", searchHandler.Search)

		// Chat sessions
		chat := api.Group("/chat")
		{
			chat.POST("/session", chatHandler.CreateSession)
			chat.GET("/session/:session_id", chatHandler.GetSession)
			chat.POST("/session/:session_id/message", chatHandler.SendMessage)
			chat.DELETE("/session/:session_id", chatHandler.DeleteSession)
		}
	}

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "Research Pro Mode API",
			"version": "1.0.0",
			"status":  "running",
		})
	})
}
