"""Base evaluator class"""
from abc import ABC, abstractmethod
from typing import Dict, Any, List
from pydantic import BaseModel


class EvaluationResult(BaseModel):
    """Standard evaluation result format"""
    evaluator_type: str
    score: float  # 0.0 to 1.0
    details: Dict[str, Any] = {}
    issues: List[Dict[str, Any]] = []
    suggestions: List[Dict[str, Any]] = []
    confidence: float = 1.0


class BaseEvaluator(ABC):
    """Base class for all evaluators"""
    
    def __init__(self, config: Dict[str, Any] = None):
        self.config = config or {}
        self.version = "1.0.0"
    
    @abstractmethod
    async def evaluate(self, conversation: Dict[str, Any]) -> EvaluationResult:
        """
        Evaluate a conversation
        
        Args:
            conversation: Full conversation data including turns, metadata, feedback
            
        Returns:
            EvaluationResult with scores, issues, and suggestions
        """
        pass
    
    @property
    @abstractmethod
    def evaluator_type(self) -> str:
        """Return the evaluator type identifier"""
        pass
    
    def extract_turns(self, conversation: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Helper to extract turns from conversation"""
        return conversation.get("turns", [])
    
    def get_metadata(self, conversation: Dict[str, Any]) -> Dict[str, Any]:
        """Helper to get conversation metadata"""
        return conversation.get("metadata", {})
