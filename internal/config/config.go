package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	// Server
	ServerHost string
	ServerPort string
	GinMode    string

	// Database
	DatabaseURL      string
	DBMaxConnections int
	DBMaxIdle        int

	// Redis
	RedisURL string

	// Python Evaluator Service
	EvaluatorServiceURL string

	// LLM
	OpenAIAPIKey     string
	AnthropicAPIKey  string
	LLMProvider      string
	LLMModel         string

	// Evaluation
	BatchSize               int
	EvaluationTimeoutSeconds int

	// Thresholds
	LatencyThresholdMS          int
	MinQualityScore             float64
	AnnotatorAgreementThreshold float64

	// Meta-Evaluation
	MetaEvalEnabled       bool
	CalibrationSampleSize int
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		// Server
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		GinMode:    getEnv("GIN_MODE", "debug"),

		// Database
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ai_agent_eval?sslmode=disable"),
		DBMaxConnections: getEnvInt("DB_MAX_CONNECTIONS", 25),
		DBMaxIdle:        getEnvInt("DB_MAX_IDLE", 10),

		// Redis
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379/0"),

		// Python Evaluator Service
		EvaluatorServiceURL: getEnv("EVALUATOR_SERVICE_URL", "http://localhost:8081"),

		// LLM
		OpenAIAPIKey:    getEnv("OPENAI_API_KEY", ""),
		AnthropicAPIKey: getEnv("ANTHROPIC_API_KEY", ""),
		LLMProvider:     getEnv("LLM_PROVIDER", "openai"),
		LLMModel:        getEnv("LLM_MODEL", "gpt-4-turbo-preview"),

		// Evaluation
		BatchSize:               getEnvInt("BATCH_SIZE", 100),
		EvaluationTimeoutSeconds: getEnvInt("EVALUATION_TIMEOUT_SECONDS", 300),

		// Thresholds
		LatencyThresholdMS:          getEnvInt("LATENCY_THRESHOLD_MS", 1000),
		MinQualityScore:             getEnvFloat("MIN_QUALITY_SCORE", 0.7),
		AnnotatorAgreementThreshold: getEnvFloat("ANNOTATOR_AGREEMENT_THRESHOLD", 0.8),

		// Meta-Evaluation
		MetaEvalEnabled:       getEnvBool("META_EVAL_ENABLED", true),
		CalibrationSampleSize: getEnvInt("CALIBRATION_SAMPLE_SIZE", 100),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
