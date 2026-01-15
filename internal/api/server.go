package api

import (
	"net/http"
	"time"

	"github.com/ai-agent-eval/internal/config"
	"github.com/ai-agent-eval/internal/queue"
	"github.com/ai-agent-eval/internal/repository"
	"github.com/ai-agent-eval/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Server represents the API server
type Server struct {
	cfg         *config.Config
	repo        *repository.Repository
	queue       *queue.RedisQueue
	evaluatorSvc *services.EvaluatorService
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, db *sqlx.DB, redisQueue *queue.RedisQueue) *Server {
	return &Server{
		cfg:         cfg,
		repo:        repository.New(db),
		queue:       redisQueue,
		evaluatorSvc: services.NewEvaluatorService(cfg.EvaluatorServiceURL),
	}
}

// Router returns the configured router
func (s *Server) Router() *gin.Engine {
	gin.SetMode(s.cfg.GinMode)
	r := gin.New()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// Health check
	r.GET("/health", s.healthCheck)

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Stats
		v1.GET("/stats", s.getStats)

		// Conversations
		v1.POST("/conversations", s.createConversation)
		v1.POST("/conversations/batch", s.batchCreateConversations)
		v1.GET("/conversations", s.listConversations)
		v1.GET("/conversations/:conversation_id", s.getConversation)

		// Feedback
		v1.POST("/feedback", s.addFeedback)

		// Evaluations
		v1.POST("/evaluations/trigger", s.triggerEvaluation)
		v1.GET("/evaluations", s.listEvaluations)
		v1.GET("/evaluations/:evaluation_id", s.getEvaluation)

		// Annotations
		v1.POST("/annotations", s.createAnnotation)
		v1.GET("/annotations/agreement/:conversation_id", s.getAnnotatorAgreement)
		v1.GET("/annotations/routing/:conversation_id", s.getRoutingDecision)

		// Improvements
		v1.POST("/improvements/analyze", s.analyzeAndGenerateSuggestions)
		v1.GET("/improvements/suggestions", s.getSuggestions)
		v1.POST("/improvements/suggestions/:suggestion_id/implement", s.markSuggestionImplemented)
		v1.GET("/improvements/patterns", s.getFailurePatterns)

		// Meta-Evaluation
		v1.POST("/meta-evaluation/calibrate", s.calibrateEvaluators)
		v1.GET("/meta-evaluation/performance", s.getEvaluatorPerformance)
	}

	return r
}

// corsMiddleware handles CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// healthCheck returns health status
// @Summary Health check
// @Description Check if the API is healthy
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
	})
}
