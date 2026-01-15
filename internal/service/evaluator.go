package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EvaluatorService handles communication with Python evaluator service
type EvaluatorService struct {
	baseURL    string
	httpClient *http.Client
}

// NewEvaluatorService creates a new evaluator service client
func NewEvaluatorService(baseURL string) *EvaluatorService {
	return &EvaluatorService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// EvaluationRequest represents a request to evaluate a conversation
type EvaluationRequest struct {
	ConversationID string                 `json:"conversation_id"`
	Turns          []map[string]interface{} `json:"turns"`
	Metadata       map[string]interface{}   `json:"metadata"`
	EvaluatorTypes []string               `json:"evaluator_types"`
}

// EvaluationResult represents the evaluation result from Python service
type EvaluationResult struct {
	EvaluationID           string                   `json:"evaluation_id"`
	ConversationID         string                   `json:"conversation_id"`
	Scores                 map[string]float64       `json:"scores"`
	ToolEvaluation         map[string]interface{}   `json:"tool_evaluation"`
	IssuesDetected         []map[string]interface{} `json:"issues_detected"`
	ImprovementSuggestions []map[string]interface{} `json:"improvement_suggestions"`
	EvaluatorVersion       string                   `json:"evaluator_version"`
	EvaluationDurationMS   int                      `json:"evaluation_duration_ms"`
}

// Evaluate sends a conversation to the Python service for evaluation
func (s *EvaluatorService) Evaluate(req *EvaluationRequest) (*EvaluationResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.httpClient.Post(
		s.baseURL+"/evaluate",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call evaluator service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("evaluator service returned status %d", resp.StatusCode)
	}

	var result EvaluationResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// AnalyzePatterns calls the Python service to analyze patterns
func (s *EvaluatorService) AnalyzePatterns(lookbackDays int) (map[string]interface{}, error) {
	resp, err := s.httpClient.Post(
		fmt.Sprintf("%s/analyze?lookback_days=%d", s.baseURL, lookbackDays),
		"application/json",
		nil,
	)
	if err != nil {
		// Return mock data if Python service is not available
		return map[string]interface{}{
			"status":                "mock",
			"analysis_period_days":  lookbackDays,
			"patterns_detected":     0,
			"suggestions_generated": 0,
			"message":               "Python evaluator service not available",
		}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// CalibrateEvaluators calls the Python service to calibrate evaluators
func (s *EvaluatorService) CalibrateEvaluators(lookbackDays int) (map[string]interface{}, error) {
	resp, err := s.httpClient.Post(
		fmt.Sprintf("%s/calibrate?lookback_days=%d", s.baseURL, lookbackDays),
		"application/json",
		nil,
	)
	if err != nil {
		// Return mock data if Python service is not available
		return map[string]interface{}{
			"status":       "mock",
			"period_days":  lookbackDays,
			"calibrations": []map[string]interface{}{},
			"message":      "Python evaluator service not available",
		}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
