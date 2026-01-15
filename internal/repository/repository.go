package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ai-agent-eval/internal/models"
	"github.com/jmoiron/sqlx"
)

// Repository provides database operations
type Repository struct {
	db *sqlx.DB
}

// New creates a new repository
func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateConversation creates a new conversation
func (r *Repository) CreateConversation(conv *models.ConversationCreate) (*models.Conversation, error) {
	turnsJSON, err := json.Marshal(conv.Turns)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal turns: %w", err)
	}

	metadataJSON := []byte("{}")
	if conv.Metadata != nil {
		metadataJSON, err = json.Marshal(conv.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO conversations (conversation_id, agent_version, turns, metadata)
		VALUES ($1, $2, $3, $4)
		RETURNING id, conversation_id, agent_version, turns, metadata, created_at, updated_at
	`

	var result models.Conversation
	err = r.db.QueryRowx(query, conv.ConversationID, conv.AgentVersion, turnsJSON, metadataJSON).
		StructScan(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Create feedback if provided
	if conv.Feedback != nil {
		if err := r.createFeedback(conv.ConversationID, conv.Feedback); err != nil {
			return nil, err
		}
	}

	return &result, nil
}

// createFeedback creates feedback for a conversation
func (r *Repository) createFeedback(conversationID string, feedback *models.Feedback) error {
	opsReviewJSON := []byte("null")
	var err error
	if feedback.OpsReview != nil {
		opsReviewJSON, err = json.Marshal(feedback.OpsReview)
		if err != nil {
			return fmt.Errorf("failed to marshal ops_review: %w", err)
		}
	}

	annotationsJSON, err := json.Marshal(feedback.Annotations)
	if err != nil {
		return fmt.Errorf("failed to marshal annotations: %w", err)
	}

	query := `
		INSERT INTO feedbacks (conversation_id, user_rating, ops_review, annotations)
		VALUES ($1, $2, $3, $4)
	`

	var userRating interface{} = nil
	if feedback.UserRating > 0 {
		userRating = feedback.UserRating
	}

	_, err = r.db.Exec(query, conversationID, userRating, opsReviewJSON, annotationsJSON)
	if err != nil {
		return fmt.Errorf("failed to create feedback: %w", err)
	}

	return nil
}

// GetConversation retrieves a conversation by ID
func (r *Repository) GetConversation(conversationID string) (*models.Conversation, error) {
	var conv models.Conversation
	query := `SELECT * FROM conversations WHERE conversation_id = $1`
	
	if err := r.db.Get(&conv, query, conversationID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return &conv, nil
}

// ListConversations lists conversations with pagination
func (r *Repository) ListConversations(agentVersion string, limit, offset int) ([]models.Conversation, error) {
	var conversations []models.Conversation
	
	query := `SELECT * FROM conversations`
	args := []interface{}{}
	argIndex := 1

	if agentVersion != "" {
		query += fmt.Sprintf(" WHERE agent_version = $%d", argIndex)
		args = append(args, agentVersion)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	if err := r.db.Select(&conversations, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}

	return conversations, nil
}

// CreateEvaluation creates an evaluation record
func (r *Repository) CreateEvaluation(eval *models.Evaluation) error {
	query := `
		INSERT INTO evaluations (
			evaluation_id, conversation_id, overall_score, response_quality_score,
			tool_accuracy_score, coherence_score, tool_evaluation, issues_detected,
			improvement_suggestions, evaluator_version, evaluation_duration_ms
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`

	return r.db.QueryRowx(
		query,
		eval.EvaluationID, eval.ConversationID, eval.OverallScore,
		eval.ResponseQualityScore, eval.ToolAccuracyScore, eval.CoherenceScore,
		eval.ToolEvaluation, eval.IssuesDetected, eval.ImprovementSuggestions,
		eval.EvaluatorVersion, eval.EvaluationDurationMS,
	).Scan(&eval.ID, &eval.CreatedAt)
}

// GetEvaluation retrieves an evaluation by ID
func (r *Repository) GetEvaluation(evaluationID string) (*models.Evaluation, error) {
	var eval models.Evaluation
	query := `SELECT * FROM evaluations WHERE evaluation_id = $1`
	
	if err := r.db.Get(&eval, query, evaluationID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get evaluation: %w", err)
	}

	return &eval, nil
}

// ListEvaluations lists evaluations with filtering
func (r *Repository) ListEvaluations(conversationID string, minScore, maxScore *float64, limit, offset int) ([]models.Evaluation, error) {
	var evaluations []models.Evaluation
	
	query := `SELECT * FROM evaluations WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if conversationID != "" {
		query += fmt.Sprintf(" AND conversation_id = $%d", argIndex)
		args = append(args, conversationID)
		argIndex++
	}

	if minScore != nil {
		query += fmt.Sprintf(" AND overall_score >= $%d", argIndex)
		args = append(args, *minScore)
		argIndex++
	}

	if maxScore != nil {
		query += fmt.Sprintf(" AND overall_score <= $%d", argIndex)
		args = append(args, *maxScore)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	if err := r.db.Select(&evaluations, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list evaluations: %w", err)
	}

	return evaluations, nil
}

// CreateAnnotation creates an annotation
func (r *Repository) CreateAnnotation(ann *models.AnnotationCreate) (*models.Annotation, error) {
	query := `
		INSERT INTO annotations (
			conversation_id, annotator_id, annotation_type, label,
			score, confidence, notes, time_spent_seconds
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, conversation_id, annotator_id, annotation_type, label,
				  score, confidence, notes, time_spent_seconds, created_at
	`

	var result models.Annotation
	err := r.db.QueryRowx(
		query,
		ann.ConversationID, ann.AnnotatorID, ann.AnnotationType, ann.Label,
		ann.Score, ann.Confidence, ann.Notes, ann.TimeSpentSeconds,
	).StructScan(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to create annotation: %w", err)
	}

	return &result, nil
}

// GetAnnotationsForConversation retrieves annotations for a conversation
func (r *Repository) GetAnnotationsForConversation(conversationID, annotationType string) ([]models.Annotation, error) {
	var annotations []models.Annotation
	
	query := `SELECT * FROM annotations WHERE conversation_id = $1`
	args := []interface{}{conversationID}

	if annotationType != "" {
		query += ` AND annotation_type = $2`
		args = append(args, annotationType)
	}

	query += ` ORDER BY created_at DESC`

	if err := r.db.Select(&annotations, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get annotations: %w", err)
	}

	return annotations, nil
}

// GetSystemStats returns system statistics
func (r *Repository) GetSystemStats() (*models.SystemStats, error) {
	stats := &models.SystemStats{}

	// Total conversations
	r.db.Get(&stats.TotalConversations, `SELECT COUNT(*) FROM conversations`)

	// Total evaluations
	r.db.Get(&stats.TotalEvaluations, `SELECT COUNT(*) FROM evaluations`)

	// Total annotations
	r.db.Get(&stats.TotalAnnotations, `SELECT COUNT(*) FROM annotations`)

	// Average quality score
	var avgScore sql.NullFloat64
	r.db.Get(&avgScore, `SELECT AVG(overall_score) FROM evaluations`)
	if avgScore.Valid {
		stats.AverageQualityScore = &avgScore.Float64
	}

	// Average user rating
	var avgRating sql.NullFloat64
	r.db.Get(&avgRating, `SELECT AVG(user_rating) FROM feedbacks WHERE user_rating IS NOT NULL`)
	if avgRating.Valid {
		stats.AverageUserRating = &avgRating.Float64
	}

	// Open issues (evaluations with issues)
	r.db.Get(&stats.OpenIssuesCount, `SELECT COUNT(*) FROM evaluations WHERE jsonb_array_length(issues_detected) > 0`)

	// Pending suggestions
	r.db.Get(&stats.PendingSuggestionsCount, `SELECT COUNT(*) FROM improvement_suggestions WHERE status = 'pending'`)

	// Evaluations in last 24h
	cutoff := time.Now().Add(-24 * time.Hour)
	r.db.Get(&stats.EvaluationsLast24H, `SELECT COUNT(*) FROM evaluations WHERE created_at >= $1`, cutoff)

	return stats, nil
}

// GetFailurePatterns retrieves failure patterns
func (r *Repository) GetFailurePatterns(resolved *bool, severity string, limit int) ([]models.FailurePattern, error) {
	var patterns []models.FailurePattern
	
	query := `SELECT * FROM failure_patterns WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if resolved != nil {
		query += fmt.Sprintf(" AND resolved = $%d", argIndex)
		args = append(args, *resolved)
		argIndex++
	}

	if severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, severity)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY occurrence_count DESC LIMIT $%d", argIndex)
	args = append(args, limit)

	if err := r.db.Select(&patterns, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get failure patterns: %w", err)
	}

	return patterns, nil
}

// GetPendingSuggestions retrieves pending suggestions
func (r *Repository) GetPendingSuggestions(minConfidence float64, suggestionType string) ([]models.StoredSuggestion, error) {
	var suggestions []models.StoredSuggestion
	
	query := `SELECT * FROM improvement_suggestions WHERE status = 'pending' AND confidence >= $1`
	args := []interface{}{minConfidence}

	if suggestionType != "" {
		query += ` AND suggestion_type = $2`
		args = append(args, suggestionType)
	}

	query += ` ORDER BY confidence DESC`

	if err := r.db.Select(&suggestions, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}

	return suggestions, nil
}

// MarkSuggestionImplemented marks a suggestion as implemented
func (r *Repository) MarkSuggestionImplemented(suggestionID string, beforeMetrics json.RawMessage) error {
	query := `
		UPDATE improvement_suggestions 
		SET status = 'implemented', implemented_at = $1, before_metrics = $2, updated_at = $1
		WHERE suggestion_id = $3
	`
	_, err := r.db.Exec(query, time.Now(), beforeMetrics, suggestionID)
	return err
}

// GetEvaluatorCalibration retrieves calibration data
func (r *Repository) GetEvaluatorCalibration(evaluatorType string) ([]models.EvaluatorCalibration, error) {
	var calibrations []models.EvaluatorCalibration
	
	query := `SELECT * FROM evaluator_calibration`
	args := []interface{}{}

	if evaluatorType != "" {
		query += ` WHERE evaluator_type = $1`
		args = append(args, evaluatorType)
	}

	query += ` ORDER BY created_at DESC`

	if err := r.db.Select(&calibrations, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get calibration: %w", err)
	}

	return calibrations, nil
}

// GetLatestEvaluationForConversation gets the latest evaluation for a conversation
func (r *Repository) GetLatestEvaluationForConversation(conversationID string) (*models.Evaluation, error) {
	var eval models.Evaluation
	query := `SELECT * FROM evaluations WHERE conversation_id = $1 ORDER BY created_at DESC LIMIT 1`
	
	if err := r.db.Get(&eval, query, conversationID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get evaluation: %w", err)
	}

	return &eval, nil
}
