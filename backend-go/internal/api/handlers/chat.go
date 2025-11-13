package handlers

import (
	"net/http"
	"time"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/agents"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/config"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/database"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatHandler struct {
	db     *gorm.DB
	cfg    *config.Config
	router *agents.RouterAgent
}

func NewChatHandler(db *gorm.DB, cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		db:     db,
		cfg:    cfg,
		router: agents.NewRouterAgent(cfg),
	}
}

func (h *ChatHandler) CreateSession(c *gin.Context) {
	var req struct {
		Mode string `json:"mode" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session := database.ChatSession{
		ID:        uuid.New().String(),
		Mode:      req.Mode,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
		Messages:  []database.Message{},
	}

	if err := h.db.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *ChatHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	var session database.ChatSession
	if err := h.db.Preload("Messages.Sources").First(&session, "id = ?", sessionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session"})
		}
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	sessionID := c.Param("session_id")

	var req struct {
		Query string `json:"query" binding:"required"`
		Mode  string `json:"mode"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get session with history
	var session database.ChatSession
	if err := h.db.Preload("Messages").First(&session, "id = ?", sessionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Save user message
	userMsg := database.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Query,
		Timestamp: time.Now().Unix(),
	}
	if err := h.db.Create(&userMsg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	// Convert history for agent processing
	conversationHistory := make([]models.Message, 0)
	for _, msg := range session.Messages {
		conversationHistory = append(conversationHistory, models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Process with agent (using mode from session or request)
	mode := session.Mode
	if req.Mode != "" {
		mode = req.Mode
	}

	startTime := time.Now()
	result, err := h.router.ProcessQueryWithContext(
		c.Request.Context(),
		req.Query,
		mode,
		conversationHistory,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save assistant message
	assistantMsg := database.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "assistant",
		Content:   result.Answer,
		Timestamp: time.Now().Unix(),
		Reasoning: result.Reasoning,
	}

	// Save sources
	for _, src := range result.Sources {
		assistantMsg.Sources = append(assistantMsg.Sources, database.Source{
			Title:       src.Title,
			URL:         src.URL,
			Snippet:     src.Snippet,
			Credibility: src.Credibility,
		})
	}

	if err := h.db.Create(&assistantMsg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save response"})
		return
	}

	// Update session timestamp
	h.db.Model(&session).Update("updated_at", time.Now().Unix())

	// Return response
	result.SessionID = sessionID
	result.ProcessingTime = time.Since(startTime).Seconds()
	result.Timestamp = time.Now().Unix()
	result.ContextUsed = len(conversationHistory) > 0

	c.JSON(http.StatusOK, result)
}

func (h *ChatHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	// Delete messages first (cascade)
	if err := h.db.Where("session_id = ?", sessionID).Delete(&database.Message{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete messages"})
		return
	}

	// Delete session
	if err := h.db.Delete(&database.ChatSession{}, "id = ?", sessionID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session deleted"})
}
