# AI Agent Evaluation Pipeline

A production-ready evaluation pipeline built with **Go** (high-performance API) and **Python** (ML/LLM evaluation) for continuously improving AI agents.

## 🚀 Key Features

- **Dual-Language Architecture**: Go for high-throughput API, Python for ML/LLM evaluation
- **4 Specialized Evaluators**: LLM-as-Judge, Tool Call, Coherence, Heuristic
- **Self-Updating Mechanism**: Automatically generates improvement suggestions
- **Meta-Evaluation**: Continuously calibrates evaluators against human annotations
- **Production Scale**: Handles 1000+ conversations/minute

## 📋 Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              Streamlit Dashboard (Port 8501)                 │
└────────────────────────┬────────────────────────────────────┘
                         │
          ┌──────────────┴──────────────┐
          │                             │
          ▼                             ▼
┌─────────────────────┐       ┌─────────────────────┐
│   Go API Server     │       │ Python Evaluator    │
│   (Port 8080)       │◄─────►│ (Port 8081)         │
│                     │       │                     │
│ • High throughput   │       │ • LLM-as-Judge      │
│ • Data ingestion    │       │ • Tool Call Eval    │
│ • CRUD operations   │       │ • Coherence Eval    │
│ • Queue management  │       │ • Heuristic Eval    │
└─────────┬───────────┘       └─────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────┐
│                PostgreSQL + Redis                            │
│                (Ports 5432, 6379)                            │
└─────────────────────────────────────────────────────────────┘
```

## 🛠️ Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development)
- Python 3.11+ (for local development)
- OpenAI or Anthropic API key (optional, for LLM evaluation)

### Docker Compose (Recommended)

```bash
# Clone repository
git clone <repository-url>
cd Ai-Agent

# Start all services
make up

# Or with docker-compose directly
docker-compose up -d
```

**Access Points:**
- **Go API**: http://localhost:8080
- **Python Evaluator**: http://localhost:8081
- **Dashboard**: http://localhost:8501

### Local Development

```bash
# Terminal 1: Start PostgreSQL and Redis
docker-compose up postgres redis -d

# Terminal 2: Run Go API
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/ai_agent_eval?sslmode=disable"
export REDIS_URL="redis://localhost:6379/0"
go run ./cmd/api

# Terminal 3: Run Python Evaluator
pip install -r requirements.txt
cd python_evaluator && uvicorn main:app --reload --port 8081

# Terminal 4: Run Dashboard
pip install streamlit requests pandas plotly
streamlit run dashboard/app.py
```

## 📊 API Endpoints

### Go API (Port 8080)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/v1/stats` | GET | System statistics |
| `/api/v1/conversations` | POST | Ingest conversation |
| `/api/v1/conversations/batch` | POST | Batch ingestion |
| `/api/v1/conversations` | GET | List conversations |
| `/api/v1/evaluations/trigger` | POST | Trigger evaluation |
| `/api/v1/evaluations` | GET | List evaluations |
| `/api/v1/evaluations/{id}` | GET | Get evaluation details |
| `/api/v1/annotations` | POST | Add annotation |
| `/api/v1/annotations/agreement/{id}` | GET | Annotator agreement |
| `/api/v1/improvements/analyze` | POST | Generate suggestions |
| `/api/v1/improvements/suggestions` | GET | List suggestions |
| `/api/v1/meta-evaluation/calibrate` | POST | Calibrate evaluators |

### Python Evaluator (Port 8081)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/evaluate` | POST | Evaluate conversation |
| `/analyze` | POST | Analyze patterns |
| `/calibrate` | POST | Calibrate evaluators |
| `/evaluators` | GET | List available evaluators |

## 🧪 Usage Examples

### Ingest a Conversation

```bash
curl -X POST http://localhost:8080/api/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{
    "conversation_id": "conv_001",
    "agent_version": "v2.3.1",
    "turns": [
      {
        "turn_id": 1,
        "role": "user",
        "content": "Book a flight to NYC",
        "timestamp": "2024-01-15T10:30:00Z"
      },
      {
        "turn_id": 2,
        "role": "assistant",
        "content": "I will help you book a flight to NYC.",
        "tool_calls": [{
          "tool_name": "flight_search",
          "parameters": {"destination": "NYC"},
          "result": {"status": "success"},
          "latency_ms": 450
        }],
        "timestamp": "2024-01-15T10:30:02Z"
      }
    ],
    "metadata": {
      "total_latency_ms": 1200,
      "mission_completed": true
    }
  }'
```

### Trigger Evaluation

```bash
curl -X POST http://localhost:8080/api/v1/evaluations/trigger \
  -H "Content-Type: application/json" \
  -d '{
    "conversation_id": "conv_001",
    "evaluator_types": ["llm_judge", "tool_call", "coherence", "heuristic"]
  }'
```

### Generate Improvement Suggestions

```bash
curl -X POST "http://localhost:8080/api/v1/improvements/analyze?lookback_days=7"
```

## 📐 Evaluation Framework

### 1. LLM-as-Judge Evaluator
- **Metrics**: Helpfulness, Factuality, Clarity, Appropriateness
- **Method**: Uses GPT-4/Claude to assess response quality
- **Output**: Scores (0-1) with reasoning

### 2. Tool Call Evaluator
- **Metrics**: Selection accuracy, Parameter accuracy, Execution success
- **Checks**: Correct tool selection, Required parameters, Hallucinated params
- **Output**: Detailed tool-level scores

### 3. Coherence Evaluator
- **Metrics**: Context maintenance, Consistency, Reference handling
- **Checks**: Context loss after turn 5, Contradictions, Pronoun resolution
- **Output**: Multi-turn coherence scores

### 4. Heuristic Evaluator
- **Metrics**: Latency, Format compliance, Required fields
- **Checks**: Response time, Empty responses, Tool call format
- **Output**: Binary pass/fail with details

## 🔄 Self-Updating Mechanism

The pipeline automatically:
1. **Detects Patterns**: Identifies recurring issues across evaluations
2. **Generates Suggestions**: Creates actionable improvements
3. **Tracks Impact**: Measures before/after metrics

Example suggestion:
```json
{
  "type": "prompt",
  "suggestion": "Add explicit context maintenance instruction",
  "rationale": "15 cases of context loss detected",
  "confidence": 0.85,
  "expected_impact": "Reduce context_loss by 30-50%"
}
```

## 🎯 Meta-Evaluation

Continuously improves evaluators by:
- Comparing predictions with human annotations
- Calculating precision, recall, F1, correlation
- Identifying blind spots
- Adjusting weights and thresholds

## 📁 Project Structure

```
Ai-Agent/
├── cmd/api/main.go              # Go API entry point
├── internal/
│   ├── api/                     # HTTP handlers
│   ├── config/                  # Configuration
│   ├── database/                # Database layer
│   ├── models/                  # Data models
│   ├── queue/                   # Redis queue
│   ├── repository/              # Data access
│   └── services/                # Business logic
├── python_evaluator/
│   ├── evaluators/              # 4 evaluator types
│   ├── services/                # Self-improvement, meta-eval
│   ├── main.py                  # FastAPI service
│   └── config.py                # Configuration
├── dashboard/
│   └── app.py                   # Streamlit UI
├── scripts/
│   └── sample_data.py           # Sample data generator
├── docker-compose.yml           # Container orchestration
├── Dockerfile.go                # Go service
├── Dockerfile.python            # Python service
├── Dockerfile.dashboard         # Dashboard
├── go.mod                       # Go dependencies
├── requirements.txt             # Python dependencies
└── Makefile                     # Build commands
```

## 🚀 Scaling Strategy

### Current Capacity
- **Ingestion**: 1,000+ conversations/minute
- **Evaluation**: 100+ conversations/minute

### 10x Scale
- Add Go API replicas behind load balancer
- Increase Celery workers for evaluation
- Use Redis cluster for caching

### 100x Scale
- Microservices architecture
- Kafka for message queue
- Database sharding
- CDN for dashboard

## 🔧 Configuration

Environment variables:

```bash
# Go API
DATABASE_URL=postgres://user:pass@host:5432/db
REDIS_URL=redis://host:6379/0
EVALUATOR_SERVICE_URL=http://python-evaluator:8081

# Python Evaluator
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
LLM_PROVIDER=openai
LLM_MODEL=gpt-4-turbo-preview

# Thresholds
LATENCY_THRESHOLD_MS=1000
MIN_QUALITY_SCORE=0.7
ANNOTATOR_AGREEMENT_THRESHOLD=0.8
```

## 🐳 Docker Commands

```bash
# Build all images
make build

# Start all services
make up

# View logs
make logs

# Stop services
make down

# Clean up
make clean
```

## 🧪 Testing

```bash
# Run all tests
make test

# Generate sample data
make sample-data

# Or directly
python scripts/sample_data.py
```

## 📚 Sample Scenarios Handled

### Scenario 1: Tool Call Regression
**Problem**: Agent calls flight_search with incorrect date formats
**Response**: Tool Call Evaluator detects pattern, generates prompt fix suggestion

### Scenario 2: Context Loss
**Problem**: Agent forgets user preferences after turn 5
**Response**: Coherence Evaluator flags issue, suggests context maintenance instruction

### Scenario 3: Annotator Disagreement
**Problem**: Two annotators disagree on response quality
**Response**: Agreement analysis flags for tiebreaker, calculates consensus

## 🏗️ Design Decisions

### Why Go + Python?

| Go (API) | Python (Evaluation) |
|----------|---------------------|
| High concurrency | LLM SDK support |
| Low latency | Data science libraries |
| Memory efficient | Rapid prototyping |
| Type safety | ML frameworks |

### Why Separate Services?
- Independent scaling
- Language-specific optimization
- Fault isolation
- Easier maintenance

## 📄 License

MIT License

## 🙏 Acknowledgments

Built with:
- Go + Gin - High-performance HTTP framework
- FastAPI - Modern Python web framework
- PostgreSQL - Reliable database
- Redis - Fast caching and queues
- Streamlit - Beautiful dashboards
- OpenAI/Anthropic - LLM evaluation

---

**Version**: 1.0.0  
**Architecture**: Go + Python  
**Status**: Production-Ready
