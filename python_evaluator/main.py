"""FastAPI service for Python Evaluator"""
from fastapi import FastAPI, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List, Dict, Any, Optional
from datetime import datetime, timedelta
import asyncio

from python_evaluator.config import get_settings
from python_evaluator.evaluators import EvaluationOrchestrator
from python_evaluator.services.self_improvement import SelfImprovementService
from python_evaluator.services.meta_evaluation import MetaEvaluationService

settings = get_settings()

app = FastAPI(
    title="AI Agent Evaluator Service",
    description="Python-based evaluation service with LLM-as-Judge and other evaluators",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Initialize services
orchestrator = EvaluationOrchestrator()


class EvaluationRequest(BaseModel):
    """Request model for evaluation"""
    conversation_id: str
    turns: List[Dict[str, Any]]
    metadata: Optional[Dict[str, Any]] = {}
    evaluator_types: Optional[List[str]] = None


class EvaluationResponse(BaseModel):
    """Response model for evaluation"""
    evaluation_id: str
    conversation_id: str
    scores: Dict[str, float]
    tool_evaluation: Optional[Dict[str, Any]] = None
    issues_detected: List[Dict[str, Any]] = []
    improvement_suggestions: List[Dict[str, Any]] = []
    evaluator_version: str
    evaluation_duration_ms: int


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat(),
        "version": "1.0.0"
    }


@app.post("/evaluate", response_model=EvaluationResponse)
async def evaluate_conversation(request: EvaluationRequest):
    """
    Evaluate a conversation using multiple evaluators
    
    - **conversation_id**: Unique identifier for the conversation
    - **turns**: List of conversation turns
    - **metadata**: Optional metadata (latency, etc.)
    - **evaluator_types**: Optional list of evaluators to run
    """
    conversation = {
        "conversation_id": request.conversation_id,
        "turns": request.turns,
        "metadata": request.metadata or {}
    }
    
    evaluator_types = request.evaluator_types
    if not evaluator_types:
        evaluator_types = ["llm_judge", "tool_call", "coherence", "heuristic"]
    
    result = await orchestrator.evaluate_conversation(conversation, evaluator_types)
    
    return EvaluationResponse(
        evaluation_id=result["evaluation_id"],
        conversation_id=result["conversation_id"],
        scores=result["scores"],
        tool_evaluation=result.get("tool_evaluation"),
        issues_detected=result["issues_detected"],
        improvement_suggestions=result["improvement_suggestions"],
        evaluator_version=result["evaluator_version"],
        evaluation_duration_ms=result["evaluation_duration_ms"]
    )


@app.post("/analyze")
async def analyze_patterns(lookback_days: int = Query(7, ge=1, le=90)):
    """
    Analyze patterns and generate improvement suggestions
    
    - **lookback_days**: Number of days to analyze
    """
    service = SelfImprovementService()
    result = await service.analyze_and_generate_suggestions(lookback_days)
    return result


@app.post("/calibrate")
async def calibrate_evaluators(lookback_days: int = Query(30, ge=7, le=180)):
    """
    Calibrate evaluators against human annotations
    
    - **lookback_days**: Calibration period in days
    """
    service = MetaEvaluationService()
    result = await service.calibrate_evaluators(lookback_days)
    return result


@app.get("/evaluators")
async def list_evaluators():
    """List available evaluators"""
    return {
        "evaluators": [
            {
                "type": "llm_judge",
                "description": "LLM-as-Judge for response quality assessment",
                "metrics": ["helpfulness", "factuality", "clarity", "appropriateness"]
            },
            {
                "type": "tool_call",
                "description": "Tool call accuracy evaluator",
                "metrics": ["selection_accuracy", "parameter_accuracy", "execution_success"]
            },
            {
                "type": "coherence",
                "description": "Multi-turn coherence evaluator",
                "metrics": ["context_maintenance", "consistency", "reference_handling", "information_retention"]
            },
            {
                "type": "heuristic",
                "description": "Fast rule-based quality checks",
                "metrics": ["latency", "format_compliance", "required_fields", "response_length"]
            }
        ]
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host=settings.service_host, port=settings.service_port)
