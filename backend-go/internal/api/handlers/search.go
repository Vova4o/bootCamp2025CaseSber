package handlers

import (
	"net/http"
	"time"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/agents"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/config"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SearchHandler struct {
	db     *gorm.DB
	cfg    *config.Config
	router *agents.RouterAgent
}

func NewSearchHandler(db *gorm.DB, cfg *config.Config) *SearchHandler {
	return &SearchHandler{
		db:     db,
		cfg:    cfg,
		router: agents.NewRouterAgent(cfg),
	}
}

func (h *SearchHandler) Search(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Route to appropriate mode
	result, err := h.router.ProcessQuery(c.Request.Context(), req.Query, req.Mode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add processing time
	result.ProcessingTime = time.Since(startTime).Seconds()
	result.Timestamp = time.Now().Unix()

	c.JSON(http.StatusOK, result)
}
