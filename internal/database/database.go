package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// New creates a new database connection
func New(databaseURL string, maxConnections, maxIdle int) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(maxIdle)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Migrate runs database migrations
func Migrate(db *sqlx.DB) error {
	migrations := []string{
		// Conversations table
		`CREATE TABLE IF NOT EXISTS conversations (
			id SERIAL PRIMARY KEY,
			conversation_id VARCHAR(255) UNIQUE NOT NULL,
			agent_version VARCHAR(100) NOT NULL,
			turns JSONB NOT NULL,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Indexes for conversations
		`CREATE INDEX IF NOT EXISTS idx_conversations_agent_version ON conversations(agent_version)`,
		`CREATE INDEX IF NOT EXISTS idx_conversations_created_at ON conversations(created_at)`,
		
		// Feedbacks table
		`CREATE TABLE IF NOT EXISTS feedbacks (
			id SERIAL PRIMARY KEY,
			conversation_id VARCHAR(255) REFERENCES conversations(conversation_id),
			user_rating INTEGER,
			ops_review JSONB,
			annotations JSONB DEFAULT '[]',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE INDEX IF NOT EXISTS idx_feedbacks_conversation_id ON feedbacks(conversation_id)`,
		
		// Evaluations table
		`CREATE TABLE IF NOT EXISTS evaluations (
			id SERIAL PRIMARY KEY,
			evaluation_id VARCHAR(255) UNIQUE NOT NULL,
			conversation_id VARCHAR(255) REFERENCES conversations(conversation_id),
			overall_score FLOAT,
			response_quality_score FLOAT,
			tool_accuracy_score FLOAT,
			coherence_score FLOAT,
			tool_evaluation JSONB DEFAULT '{}',
			issues_detected JSONB DEFAULT '[]',
			improvement_suggestions JSONB DEFAULT '[]',
			evaluator_version VARCHAR(50),
			evaluation_duration_ms INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE INDEX IF NOT EXISTS idx_evaluations_conversation_id ON evaluations(conversation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_evaluations_overall_score ON evaluations(overall_score)`,
		`CREATE INDEX IF NOT EXISTS idx_evaluations_created_at ON evaluations(created_at)`,
		
		// Annotations table
		`CREATE TABLE IF NOT EXISTS annotations (
			id SERIAL PRIMARY KEY,
			conversation_id VARCHAR(255),
			annotator_id VARCHAR(255) NOT NULL,
			annotation_type VARCHAR(100) NOT NULL,
			label VARCHAR(255) NOT NULL,
			score FLOAT,
			confidence FLOAT,
			notes TEXT,
			time_spent_seconds INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE INDEX IF NOT EXISTS idx_annotations_conversation_id ON annotations(conversation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_annotations_annotator_id ON annotations(annotator_id)`,
		`CREATE INDEX IF NOT EXISTS idx_annotations_type ON annotations(annotation_type)`,
		
		// Annotator Performance table
		`CREATE TABLE IF NOT EXISTS annotator_performance (
			id SERIAL PRIMARY KEY,
			annotator_id VARCHAR(255) UNIQUE NOT NULL,
			total_annotations INTEGER DEFAULT 0,
			agreement_rate FLOAT,
			consistency_score FLOAT,
			accuracy_vs_ground_truth FLOAT,
			specializations JSONB DEFAULT '[]',
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Failure Patterns table
		`CREATE TABLE IF NOT EXISTS failure_patterns (
			id SERIAL PRIMARY KEY,
			pattern_id VARCHAR(255) UNIQUE NOT NULL,
			pattern_type VARCHAR(100) NOT NULL,
			description TEXT NOT NULL,
			severity VARCHAR(50) NOT NULL,
			first_seen TIMESTAMP NOT NULL,
			last_seen TIMESTAMP NOT NULL,
			occurrence_count INTEGER DEFAULT 1,
			affected_versions JSONB DEFAULT '[]',
			example_conversations JSONB DEFAULT '[]',
			resolved BOOLEAN DEFAULT FALSE,
			resolution_notes TEXT,
			related_suggestion_id VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE INDEX IF NOT EXISTS idx_failure_patterns_type ON failure_patterns(pattern_type)`,
		`CREATE INDEX IF NOT EXISTS idx_failure_patterns_severity ON failure_patterns(severity)`,
		`CREATE INDEX IF NOT EXISTS idx_failure_patterns_resolved ON failure_patterns(resolved)`,
		
		// Improvement Suggestions table
		`CREATE TABLE IF NOT EXISTS improvement_suggestions (
			id SERIAL PRIMARY KEY,
			suggestion_id VARCHAR(255) UNIQUE NOT NULL,
			suggestion_type VARCHAR(100) NOT NULL,
			suggestion TEXT NOT NULL,
			rationale TEXT NOT NULL,
			confidence FLOAT NOT NULL,
			pattern_detected JSONB DEFAULT '{}',
			affected_conversations JSONB DEFAULT '[]',
			frequency INTEGER DEFAULT 1,
			status VARCHAR(50) DEFAULT 'pending',
			implemented_at TIMESTAMP,
			impact_measured BOOLEAN DEFAULT FALSE,
			before_metrics JSONB,
			after_metrics JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE INDEX IF NOT EXISTS idx_suggestions_type ON improvement_suggestions(suggestion_type)`,
		`CREATE INDEX IF NOT EXISTS idx_suggestions_status ON improvement_suggestions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_suggestions_confidence ON improvement_suggestions(confidence)`,
		
		// Evaluator Calibration table
		`CREATE TABLE IF NOT EXISTS evaluator_calibration (
			id SERIAL PRIMARY KEY,
			evaluator_type VARCHAR(100) NOT NULL,
			evaluator_version VARCHAR(50),
			precision FLOAT,
			recall FLOAT,
			f1_score FLOAT,
			correlation_with_human FLOAT,
			calibration_samples INTEGER DEFAULT 0,
			false_positive_rate FLOAT,
			false_negative_rate FLOAT,
			missed_patterns JSONB DEFAULT '[]',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE INDEX IF NOT EXISTS idx_calibration_evaluator_type ON evaluator_calibration(evaluator_type)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}
