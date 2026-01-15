package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// ToolCall represents a tool call made by the agent
type ToolCall struct {
	ToolName   string                 `json:"tool_name"`
	Parameters map[string]interface{} `json:"parameters"`
	Result     map[string]interface{} `json:"result,omitempty"`
	LatencyMS  int                    `json:"latency_ms,omitempty"`
}

// Turn represents a single turn in a conversation
type Turn struct {
	TurnID    int        `json:"turn_id"`
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// OpsReview represents an operations review
type OpsReview struct {
	Quality string `json:"quality"`
	Notes   string `json:"notes,omitempty"`
}

// AnnotationItem represents a single annotation
type AnnotationItem struct {
	Type        string  `json:"type"`
	Label       string  `json:"label"`
	AnnotatorID string  `json:"annotator_id"`
	Score       float64 `json:"score,omitempty"`
	Confidence  float64 `json:"confidence,omitempty"`
}

// Feedback represents feedback data
type Feedback struct {
	UserRating  int              `json:"user_rating,omitempty"`
	OpsReview   *OpsReview       `json:"ops_review,omitempty"`
	Annotations []AnnotationItem `json:"annotations,omitempty"`
}

// ConversationMetadata represents conversation metadata
type ConversationMetadata struct {
	TotalLatencyMS   int  `json:"total_latency_ms,omitempty"`
	MissionCompleted bool `json:"mission_completed,omitempty"`
}

// Conversation represents a conversation to be evaluated
type Conversation struct {
	ID             int64                `json:"id" db:"id"`
	ConversationID string               `json:"conversation_id" db:"conversation_id"`
	AgentVersion   string               `json:"agent_version" db:"agent_version"`
	Turns          json.RawMessage      `json:"turns" db:"turns"`
	Metadata       json.RawMessage      `json:"metadata" db:"metadata"`
	CreatedAt      time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at" db:"updated_at"`
}

// ConversationCreate represents the input for creating a conversation
type ConversationCreate struct {
	ConversationID string               `json:"conversation_id" binding:"required"`
	AgentVersion   string               `json:"agent_version" binding:"required"`
	Turns          []Turn               `json:"turns" binding:"required,min=1"`
	Feedback       *Feedback            `json:"feedback,omitempty"`
	Metadata       *ConversationMetadata `json:"metadata,omitempty"`
}

// EvaluationScores represents evaluation scores
type EvaluationScores struct {
	Overall         float64 `json:"overall"`
	ResponseQuality float64 `json:"response_quality"`
	ToolAccuracy    float64 `json:"tool_accuracy"`
	Coherence       float64 `json:"coherence"`
}

// ToolEvaluation represents tool-specific evaluation
type ToolEvaluation struct {
	SelectionAccuracy    float64  `json:"selection_accuracy"`
	ParameterAccuracy    float64  `json:"parameter_accuracy"`
	ExecutionSuccess     bool     `json:"execution_success"`
	HallucinatedParams   []string `json:"hallucinated_parameters,omitempty"`
}

// IssueDetected represents a detected issue
type IssueDetected struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	TurnID      int    `json:"turn_id,omitempty"`
}

// ImprovementSuggestion represents an improvement suggestion
type ImprovementSuggestion struct {
	Type           string  `json:"type"`
	Suggestion     string  `json:"suggestion"`
	Rationale      string  `json:"rationale"`
	Confidence     float64 `json:"confidence"`
	ExpectedImpact string  `json:"expected_impact,omitempty"`
}

// Evaluation represents an evaluation result
type Evaluation struct {
	ID                     int64           `json:"id" db:"id"`
	EvaluationID           string          `json:"evaluation_id" db:"evaluation_id"`
	ConversationID         string          `json:"conversation_id" db:"conversation_id"`
	OverallScore           float64         `json:"overall_score" db:"overall_score"`
	ResponseQualityScore   float64         `json:"response_quality_score" db:"response_quality_score"`
	ToolAccuracyScore      float64         `json:"tool_accuracy_score" db:"tool_accuracy_score"`
	CoherenceScore         float64         `json:"coherence_score" db:"coherence_score"`
	ToolEvaluation         json.RawMessage `json:"tool_evaluation" db:"tool_evaluation"`
	IssuesDetected         json.RawMessage `json:"issues_detected" db:"issues_detected"`
	ImprovementSuggestions json.RawMessage `json:"improvement_suggestions" db:"improvement_suggestions"`
	EvaluatorVersion       string          `json:"evaluator_version" db:"evaluator_version"`
	EvaluationDurationMS   int             `json:"evaluation_duration_ms" db:"evaluation_duration_ms"`
	CreatedAt              time.Time       `json:"created_at" db:"created_at"`
}

// EvaluationResponse represents the full evaluation response
type EvaluationResponse struct {
	EvaluationID           string                  `json:"evaluation_id"`
	ConversationID         string                  `json:"conversation_id"`
	Scores                 EvaluationScores        `json:"scores"`
	ToolEvaluation         *ToolEvaluation         `json:"tool_evaluation,omitempty"`
	IssuesDetected         []IssueDetected         `json:"issues_detected"`
	ImprovementSuggestions []ImprovementSuggestion `json:"improvement_suggestions"`
	EvaluationDurationMS   int                     `json:"evaluation_duration_ms,omitempty"`
	CreatedAt              time.Time               `json:"created_at"`
}

// FeedbackRecord represents stored feedback
type FeedbackRecord struct {
	ID             int64           `json:"id" db:"id"`
	ConversationID string          `json:"conversation_id" db:"conversation_id"`
	UserRating     sql.NullInt32   `json:"user_rating" db:"user_rating"`
	OpsReview      json.RawMessage `json:"ops_review" db:"ops_review"`
	Annotations    json.RawMessage `json:"annotations" db:"annotations"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

// Annotation represents a human annotation
type Annotation struct {
	ID               int64           `json:"id" db:"id"`
	ConversationID   string          `json:"conversation_id" db:"conversation_id"`
	AnnotatorID      string          `json:"annotator_id" db:"annotator_id"`
	AnnotationType   string          `json:"annotation_type" db:"annotation_type"`
	Label            string          `json:"label" db:"label"`
	Score            sql.NullFloat64 `json:"score" db:"score"`
	Confidence       sql.NullFloat64 `json:"confidence" db:"confidence"`
	Notes            sql.NullString  `json:"notes" db:"notes"`
	TimeSpentSeconds sql.NullInt32   `json:"time_spent_seconds" db:"time_spent_seconds"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
}

// AnnotationCreate represents input for creating annotation
type AnnotationCreate struct {
	ConversationID   string   `json:"conversation_id" binding:"required"`
	AnnotatorID      string   `json:"annotator_id" binding:"required"`
	AnnotationType   string   `json:"annotation_type" binding:"required"`
	Label            string   `json:"label" binding:"required"`
	Score            *float64 `json:"score,omitempty"`
	Confidence       *float64 `json:"confidence,omitempty"`
	Notes            string   `json:"notes,omitempty"`
	TimeSpentSeconds int      `json:"time_spent_seconds,omitempty"`
}

// FailurePattern represents a detected failure pattern
type FailurePattern struct {
	ID                   int64           `json:"id" db:"id"`
	PatternID            string          `json:"pattern_id" db:"pattern_id"`
	PatternType          string          `json:"pattern_type" db:"pattern_type"`
	Description          string          `json:"description" db:"description"`
	Severity             string          `json:"severity" db:"severity"`
	FirstSeen            time.Time       `json:"first_seen" db:"first_seen"`
	LastSeen             time.Time       `json:"last_seen" db:"last_seen"`
	OccurrenceCount      int             `json:"occurrence_count" db:"occurrence_count"`
	AffectedVersions     json.RawMessage `json:"affected_versions" db:"affected_versions"`
	ExampleConversations json.RawMessage `json:"example_conversations" db:"example_conversations"`
	Resolved             bool            `json:"resolved" db:"resolved"`
	ResolutionNotes      sql.NullString  `json:"resolution_notes" db:"resolution_notes"`
	RelatedSuggestionID  sql.NullString  `json:"related_suggestion_id" db:"related_suggestion_id"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

// StoredSuggestion represents a stored improvement suggestion
type StoredSuggestion struct {
	ID                    int64           `json:"id" db:"id"`
	SuggestionID          string          `json:"suggestion_id" db:"suggestion_id"`
	SuggestionType        string          `json:"suggestion_type" db:"suggestion_type"`
	Suggestion            string          `json:"suggestion" db:"suggestion"`
	Rationale             string          `json:"rationale" db:"rationale"`
	Confidence            float64         `json:"confidence" db:"confidence"`
	PatternDetected       json.RawMessage `json:"pattern_detected" db:"pattern_detected"`
	AffectedConversations json.RawMessage `json:"affected_conversations" db:"affected_conversations"`
	Frequency             int             `json:"frequency" db:"frequency"`
	Status                string          `json:"status" db:"status"`
	ImplementedAt         sql.NullTime    `json:"implemented_at" db:"implemented_at"`
	ImpactMeasured        bool            `json:"impact_measured" db:"impact_measured"`
	BeforeMetrics         json.RawMessage `json:"before_metrics" db:"before_metrics"`
	AfterMetrics          json.RawMessage `json:"after_metrics" db:"after_metrics"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at" db:"updated_at"`
}

// EvaluatorCalibration represents evaluator calibration data
type EvaluatorCalibration struct {
	ID                  int64           `json:"id" db:"id"`
	EvaluatorType       string          `json:"evaluator_type" db:"evaluator_type"`
	EvaluatorVersion    string          `json:"evaluator_version" db:"evaluator_version"`
	Precision           sql.NullFloat64 `json:"precision" db:"precision"`
	Recall              sql.NullFloat64 `json:"recall" db:"recall"`
	F1Score             sql.NullFloat64 `json:"f1_score" db:"f1_score"`
	CorrelationWithHuman sql.NullFloat64 `json:"correlation_with_human" db:"correlation_with_human"`
	CalibrationSamples  int             `json:"calibration_samples" db:"calibration_samples"`
	FalsePositiveRate   sql.NullFloat64 `json:"false_positive_rate" db:"false_positive_rate"`
	FalseNegativeRate   sql.NullFloat64 `json:"false_negative_rate" db:"false_negative_rate"`
	MissedPatterns      json.RawMessage `json:"missed_patterns" db:"missed_patterns"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

// SystemStats represents system statistics
type SystemStats struct {
	TotalConversations      int      `json:"total_conversations"`
	TotalEvaluations        int      `json:"total_evaluations"`
	TotalAnnotations        int      `json:"total_annotations"`
	AverageQualityScore     *float64 `json:"average_quality_score"`
	AverageUserRating       *float64 `json:"average_user_rating"`
	OpenIssuesCount         int      `json:"open_issues_count"`
	PendingSuggestionsCount int      `json:"pending_suggestions_count"`
	EvaluationsLast24H      int      `json:"evaluations_last_24h"`
}

// AnnotatorAgreement represents agreement analysis result
type AnnotatorAgreement struct {
	ConversationID        string        `json:"conversation_id"`
	AnnotationType        string        `json:"annotation_type"`
	Annotators            []string      `json:"annotators"`
	AgreementScore        float64       `json:"agreement_score"`
	MajorityLabel         string        `json:"majority_label,omitempty"`
	NeedsTiebreaker       bool          `json:"needs_tiebreaker"`
	IndividualAnnotations []Annotation  `json:"individual_annotations"`
}

// RoutingDecision represents routing decision for human review
type RoutingDecision struct {
	ConversationID         string   `json:"conversation_id"`
	NeedsHumanReview       bool     `json:"needs_human_review"`
	Priority               string   `json:"priority"`
	RoutingReason          []string `json:"routing_reason"`
	AutoLabel              bool     `json:"auto_label"`
	SuggestedAnnotationTypes []string `json:"suggested_annotation_types"`
}

// EvaluationRequest represents a request to evaluate
type EvaluationRequest struct {
	ConversationID string   `json:"conversation_id" binding:"required"`
	EvaluatorTypes []string `json:"evaluator_types,omitempty"`
}

// BatchIngestResponse represents batch ingestion response
type BatchIngestResponse struct {
	Ingested        int      `json:"ingested"`
	ConversationIDs []string `json:"conversation_ids"`
}
