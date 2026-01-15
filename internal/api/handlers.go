package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ai-agent-eval/internal/models"
	"github.com/ai-agent-eval/internal/queue"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getStats returns system statistics
// @Summary Get system statistics
// @Tags Analytics
// @Produce json
// @Success 200 {object} models.SystemStats
// @Router /api/v1/stats [get]
func (s *Server) getStats(c *gin.Context) {
	stats, err := s.repo.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// createConversation ingests a new conversation
// @Summary Ingest a conversation
// @Tags Ingestion
// @Accept json
// @Produce json
// @Param conversation body models.ConversationCreate true "Conversation data"
// @Param auto_evaluate query bool false "Auto trigger evaluation" default(true)
// @Success 201 {object} models.Conversation
// @Router /api/v1/conversations [post]
func (s *Server) createConversation(c *gin.Context) {
	var conv models.ConversationCreate
	if err := c.ShouldBindJSON(&conv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := s.repo.CreateConversation(&conv)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Auto evaluate if requested
	autoEvaluate := c.DefaultQuery("auto_evaluate", "true") == "true"
	if autoEvaluate {
		task := &queue.Task{
			ID:             uuid.New().String(),
			Type:           "evaluate",
			ConversationID: conv.ConversationID,
			EvaluatorTypes: []string{"llm_judge", "tool_call", "coherence", "heuristic"},
			CreatedAt:      time.Now(),
		}
		if err := s.queue.Enqueue("evaluations", task); err != nil {
			// Log but don't fail
			_ = err
		}
	}

	c.JSON(http.StatusCreated, created)
}

// batchCreateConversations ingests multiple conversations
// @Summary Batch ingest conversations
// @Tags Ingestion
// @Accept json
// @Produce json
// @Param conversations body []models.ConversationCreate true "Conversations data"
// @Param auto_evaluate query bool false "Auto trigger evaluation" default(true)
// @Success 201 {object} models.BatchIngestResponse
// @Router /api/v1/conversations/batch [post]
func (s *Server) batchCreateConversations(c *gin.Context) {
	var convs []models.ConversationCreate
	if err := c.ShouldBindJSON(&convs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversationIDs := make([]string, 0, len(convs))
	autoEvaluate := c.DefaultQuery("auto_evaluate", "true") == "true"

	for _, conv := range convs {
		_, err := s.repo.CreateConversation(&conv)
		if err != nil {
			continue // Skip failed ones
		}
		conversationIDs = append(conversationIDs, conv.ConversationID)

		if autoEvaluate {
			task := &queue.Task{
				ID:             uuid.New().String(),
				Type:           "evaluate",
				ConversationID: conv.ConversationID,
				EvaluatorTypes: []string{"llm_judge", "tool_call", "coherence", "heuristic"},
				CreatedAt:      time.Now(),
			}
			_ = s.queue.Enqueue("evaluations", task)
		}
	}

	c.JSON(http.StatusCreated, models.BatchIngestResponse{
		Ingested:        len(conversationIDs),
		ConversationIDs: conversationIDs,
	})
}

// listConversations lists conversations
// @Summary List conversations
// @Tags Query
// @Produce json
// @Param agent_version query string false "Filter by agent version"
// @Param limit query int false "Limit" default(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/conversations [get]
func (s *Server) listConversations(c *gin.Context) {
	agentVersion := c.Query("agent_version")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	convs, err := s.repo.ListConversations(agentVersion, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversations": convs,
		"count":         len(convs),
	})
}

// getConversation retrieves a conversation by ID
// @Summary Get conversation
// @Tags Query
// @Produce json
// @Param conversation_id path string true "Conversation ID"
// @Success 200 {object} models.Conversation
// @Router /api/v1/conversations/{conversation_id} [get]
func (s *Server) getConversation(c *gin.Context) {
	conversationID := c.Param("conversation_id")

	conv, err := s.repo.GetConversation(conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if conv == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// addFeedback adds feedback to a conversation
// @Summary Add feedback
// @Tags Ingestion
// @Accept json
// @Produce json
// @Param feedback body models.Feedback true "Feedback data"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/feedback [post]
func (s *Server) addFeedback(c *gin.Context) {
	var req struct {
		ConversationID string          `json:"conversation_id" binding:"required"`
		Feedback       models.Feedback `json:"feedback" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// This would normally add to the feedbacks table
	c.JSON(http.StatusCreated, gin.H{
		"status":          "success",
		"conversation_id": req.ConversationID,
	})
}

// triggerEvaluation triggers an evaluation
// @Summary Trigger evaluation
// @Tags Evaluation
// @Accept json
// @Produce json
// @Param request body models.EvaluationRequest true "Evaluation request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/evaluations/trigger [post]
func (s *Server) triggerEvaluation(c *gin.Context) {
	var req models.EvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if conversation exists
	conv, err := s.repo.GetConversation(req.ConversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if conv == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	// Default evaluator types
	evaluatorTypes := req.EvaluatorTypes
	if len(evaluatorTypes) == 0 {
		evaluatorTypes = []string{"llm_judge", "tool_call", "coherence", "heuristic"}
	}

	// Queue the evaluation
	taskID := uuid.New().String()
	task := &queue.Task{
		ID:             taskID,
		Type:           "evaluate",
		ConversationID: req.ConversationID,
		EvaluatorTypes: evaluatorTypes,
		CreatedAt:      time.Now(),
	}

	if err := s.queue.Enqueue("evaluations", task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue evaluation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":         taskID,
		"conversation_id": req.ConversationID,
		"status":          "queued",
	})
}

// listEvaluations lists evaluations
// @Summary List evaluations
// @Tags Evaluation
// @Produce json
// @Param conversation_id query string false "Filter by conversation ID"
// @Param min_score query number false "Minimum overall score"
// @Param max_score query number false "Maximum overall score"
// @Param limit query int false "Limit" default(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/evaluations [get]
func (s *Server) listEvaluations(c *gin.Context) {
	conversationID := c.Query("conversation_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var minScore, maxScore *float64
	if min := c.Query("min_score"); min != "" {
		if v, err := strconv.ParseFloat(min, 64); err == nil {
			minScore = &v
		}
	}
	if max := c.Query("max_score"); max != "" {
		if v, err := strconv.ParseFloat(max, 64); err == nil {
			maxScore = &v
		}
	}

	evals, err := s.repo.ListEvaluations(conversationID, minScore, maxScore, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	results := make([]gin.H, 0, len(evals))
	for _, e := range evals {
		results = append(results, gin.H{
			"evaluation_id":   e.EvaluationID,
			"conversation_id": e.ConversationID,
			"overall_score":   e.OverallScore,
			"created_at":      e.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"evaluations": results,
		"count":       len(results),
	})
}

// getEvaluation retrieves an evaluation by ID
// @Summary Get evaluation
// @Tags Evaluation
// @Produce json
// @Param evaluation_id path string true "Evaluation ID"
// @Success 200 {object} models.EvaluationResponse
// @Router /api/v1/evaluations/{evaluation_id} [get]
func (s *Server) getEvaluation(c *gin.Context) {
	evaluationID := c.Param("evaluation_id")

	eval, err := s.repo.GetEvaluation(evaluationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if eval == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evaluation not found"})
		return
	}

	// Parse JSON fields
	var toolEval models.ToolEvaluation
	var issues []models.IssueDetected
	var suggestions []models.ImprovementSuggestion

	json.Unmarshal(eval.ToolEvaluation, &toolEval)
	json.Unmarshal(eval.IssuesDetected, &issues)
	json.Unmarshal(eval.ImprovementSuggestions, &suggestions)

	response := models.EvaluationResponse{
		EvaluationID:   eval.EvaluationID,
		ConversationID: eval.ConversationID,
		Scores: models.EvaluationScores{
			Overall:         eval.OverallScore,
			ResponseQuality: eval.ResponseQualityScore,
			ToolAccuracy:    eval.ToolAccuracyScore,
			Coherence:       eval.CoherenceScore,
		},
		ToolEvaluation:         &toolEval,
		IssuesDetected:         issues,
		ImprovementSuggestions: suggestions,
		EvaluationDurationMS:   eval.EvaluationDurationMS,
		CreatedAt:              eval.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// createAnnotation creates a new annotation
// @Summary Create annotation
// @Tags Annotations
// @Accept json
// @Produce json
// @Param annotation body models.AnnotationCreate true "Annotation data"
// @Success 201 {object} models.Annotation
// @Router /api/v1/annotations [post]
func (s *Server) createAnnotation(c *gin.Context) {
	var ann models.AnnotationCreate
	if err := c.ShouldBindJSON(&ann); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := s.repo.CreateAnnotation(&ann)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// getAnnotatorAgreement analyzes annotator agreement
// @Summary Get annotator agreement
// @Tags Annotations
// @Produce json
// @Param conversation_id path string true "Conversation ID"
// @Param annotation_type query string true "Annotation type"
// @Success 200 {object} models.AnnotatorAgreement
// @Router /api/v1/annotations/agreement/{conversation_id} [get]
func (s *Server) getAnnotatorAgreement(c *gin.Context) {
	conversationID := c.Param("conversation_id")
	annotationType := c.Query("annotation_type")

	if annotationType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "annotation_type is required"})
		return
	}

	annotations, err := s.repo.GetAnnotationsForConversation(conversationID, annotationType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate agreement
	annotators := make([]string, 0)
	labelCounts := make(map[string]int)

	for _, ann := range annotations {
		annotators = append(annotators, ann.AnnotatorID)
		labelCounts[ann.Label]++
	}

	// Find majority label and agreement
	var majorityLabel string
	maxCount := 0
	for label, count := range labelCounts {
		if count > maxCount {
			maxCount = count
			majorityLabel = label
		}
	}

	agreementScore := 1.0
	if len(annotations) > 1 {
		agreementScore = float64(maxCount) / float64(len(annotations))
	}

	needsTiebreaker := agreementScore < s.cfg.AnnotatorAgreementThreshold

	c.JSON(http.StatusOK, models.AnnotatorAgreement{
		ConversationID:        conversationID,
		AnnotationType:        annotationType,
		Annotators:            annotators,
		AgreementScore:        agreementScore,
		MajorityLabel:         majorityLabel,
		NeedsTiebreaker:       needsTiebreaker,
		IndividualAnnotations: annotations,
	})
}

// getRoutingDecision returns routing decision for a conversation
// @Summary Get routing decision
// @Tags Annotations
// @Produce json
// @Param conversation_id path string true "Conversation ID"
// @Success 200 {object} models.RoutingDecision
// @Router /api/v1/annotations/routing/{conversation_id} [get]
func (s *Server) getRoutingDecision(c *gin.Context) {
	conversationID := c.Param("conversation_id")

	eval, err := s.repo.GetLatestEvaluationForConversation(conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if eval == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No evaluation found for conversation"})
		return
	}

	// Parse issues
	var issues []models.IssueDetected
	json.Unmarshal(eval.IssuesDetected, &issues)

	// Determine routing
	needsReview := false
	routingReason := []string{}
	priority := "low"

	if eval.OverallScore < 0.4 {
		needsReview = true
		routingReason = append(routingReason, "Low quality score")
		priority = "high"
	}

	criticalCount := 0
	for _, issue := range issues {
		if issue.Severity == "critical" {
			criticalCount++
		}
	}

	if criticalCount > 0 {
		needsReview = true
		routingReason = append(routingReason, "Critical issues detected")
		priority = "high"
	}

	suggestedTypes := []string{"general_quality"}
	for _, issue := range issues {
		if issue.Type == "tool" || issue.Type == "tool_execution_failure" {
			suggestedTypes = append(suggestedTypes, "tool_accuracy")
		}
		if issue.Type == "context_loss" || issue.Type == "coherence" {
			suggestedTypes = append(suggestedTypes, "coherence")
		}
	}

	c.JSON(http.StatusOK, models.RoutingDecision{
		ConversationID:           conversationID,
		NeedsHumanReview:         needsReview,
		Priority:                 priority,
		RoutingReason:            routingReason,
		AutoLabel:                !needsReview,
		SuggestedAnnotationTypes: suggestedTypes,
	})
}

// analyzeAndGenerateSuggestions triggers analysis
// @Summary Analyze and generate suggestions
// @Tags Self-Improvement
// @Produce json
// @Param lookback_days query int false "Days to analyze" default(7)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/improvements/analyze [post]
func (s *Server) analyzeAndGenerateSuggestions(c *gin.Context) {
	lookbackDays, _ := strconv.Atoi(c.DefaultQuery("lookback_days", "7"))

	// Call Python evaluator service for analysis
	result, err := s.evaluatorSvc.AnalyzePatterns(lookbackDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// getSuggestions returns improvement suggestions
// @Summary Get improvement suggestions
// @Tags Self-Improvement
// @Produce json
// @Param min_confidence query number false "Minimum confidence" default(0.7)
// @Param suggestion_type query string false "Filter by type"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/improvements/suggestions [get]
func (s *Server) getSuggestions(c *gin.Context) {
	minConfidence, _ := strconv.ParseFloat(c.DefaultQuery("min_confidence", "0.7"), 64)
	suggestionType := c.Query("suggestion_type")

	suggestions, err := s.repo.GetPendingSuggestions(minConfidence, suggestionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suggestions": suggestions,
		"count":       len(suggestions),
	})
}

// markSuggestionImplemented marks a suggestion as implemented
// @Summary Mark suggestion implemented
// @Tags Self-Improvement
// @Produce json
// @Param suggestion_id path string true "Suggestion ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/improvements/suggestions/{suggestion_id}/implement [post]
func (s *Server) markSuggestionImplemented(c *gin.Context) {
	suggestionID := c.Param("suggestion_id")

	var req struct {
		BeforeMetrics map[string]interface{} `json:"before_metrics"`
	}
	c.ShouldBindJSON(&req)

	beforeMetrics, _ := json.Marshal(req.BeforeMetrics)

	if err := s.repo.MarkSuggestionImplemented(suggestionID, beforeMetrics); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"suggestion_id": suggestionID,
	})
}

// getFailurePatterns returns failure patterns
// @Summary Get failure patterns
// @Tags Self-Improvement
// @Produce json
// @Param resolved query bool false "Filter by resolved status"
// @Param severity query string false "Filter by severity"
// @Param limit query int false "Limit" default(50)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/improvements/patterns [get]
func (s *Server) getFailurePatterns(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	severity := c.Query("severity")

	var resolved *bool
	if r := c.Query("resolved"); r != "" {
		v := r == "true"
		resolved = &v
	}

	patterns, err := s.repo.GetFailurePatterns(resolved, severity, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"patterns": patterns,
		"count":    len(patterns),
	})
}

// calibrateEvaluators triggers evaluator calibration
// @Summary Calibrate evaluators
// @Tags Meta-Evaluation
// @Produce json
// @Param lookback_days query int false "Days to analyze" default(30)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/meta-evaluation/calibrate [post]
func (s *Server) calibrateEvaluators(c *gin.Context) {
	lookbackDays, _ := strconv.Atoi(c.DefaultQuery("lookback_days", "30"))

	// Call Python evaluator service for calibration
	result, err := s.evaluatorSvc.CalibrateEvaluators(lookbackDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// getEvaluatorPerformance returns evaluator performance metrics
// @Summary Get evaluator performance
// @Tags Meta-Evaluation
// @Produce json
// @Param evaluator_type query string false "Filter by evaluator type"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/meta-evaluation/performance [get]
func (s *Server) getEvaluatorPerformance(c *gin.Context) {
	evaluatorType := c.Query("evaluator_type")

	calibrations, err := s.repo.GetEvaluatorCalibration(evaluatorType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"evaluators": calibrations,
		"count":      len(calibrations),
	})
}
