"""Orchestrator to coordinate all evaluators"""
import asyncio
from typing import Dict, Any, List
from datetime import datetime
import uuid
from python_evaluator.evaluators.base import BaseEvaluator, EvaluationResult
from python_evaluator.evaluators.llm_judge import LLMJudgeEvaluator
from python_evaluator.evaluators.tool_call import ToolCallEvaluator
from python_evaluator.evaluators.coherence import CoherenceEvaluator
from python_evaluator.evaluators.heuristic import HeuristicEvaluator


class EvaluationOrchestrator:
    """Orchestrates multiple evaluators and combines results"""
    
    def __init__(self):
        self.evaluators: Dict[str, BaseEvaluator] = {
            "llm_judge": LLMJudgeEvaluator(),
            "tool_call": ToolCallEvaluator(),
            "coherence": CoherenceEvaluator(),
            "heuristic": HeuristicEvaluator()
        }
    
    async def evaluate_conversation(
        self,
        conversation: Dict[str, Any],
        evaluator_types: List[str] = None
    ) -> Dict[str, Any]:
        """
        Run multiple evaluators on a conversation
        
        Args:
            conversation: Conversation data
            evaluator_types: List of evaluator types to run (default: all)
            
        Returns:
            Combined evaluation results
        """
        start_time = datetime.now()
        
        if evaluator_types is None:
            evaluator_types = list(self.evaluators.keys())
        
        tasks = []
        for eval_type in evaluator_types:
            if eval_type in self.evaluators:
                evaluator = self.evaluators[eval_type]
                tasks.append(evaluator.evaluate(conversation))
        
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        evaluation_results = []
        for eval_type, result in zip(evaluator_types, results):
            if isinstance(result, Exception):
                evaluation_results.append(EvaluationResult(
                    evaluator_type=eval_type,
                    score=0.5,
                    details={"error": str(result)},
                    confidence=0.0
                ))
            else:
                evaluation_results.append(result)
        
        combined = self._combine_results(evaluation_results)
        
        duration_ms = int((datetime.now() - start_time).total_seconds() * 1000)
        combined["evaluation_duration_ms"] = duration_ms
        combined["evaluation_id"] = f"eval_{uuid.uuid4().hex[:12]}"
        combined["conversation_id"] = conversation.get("conversation_id", "unknown")
        
        return combined
    
    def _combine_results(self, results: List[EvaluationResult]) -> Dict[str, Any]:
        """Combine results from multiple evaluators"""
        scores = {}
        all_issues = []
        all_suggestions = []
        tool_evaluation = None
        
        for result in results:
            if result.evaluator_type == "llm_judge":
                scores["response_quality"] = result.score
            elif result.evaluator_type == "tool_call":
                scores["tool_accuracy"] = result.score
                tool_evaluation = result.details
            elif result.evaluator_type == "coherence":
                scores["coherence"] = result.score
            elif result.evaluator_type == "heuristic":
                scores["heuristic"] = result.score
            
            all_issues.extend(result.issues)
            all_suggestions.extend(result.suggestions)
        
        weights = {
            "response_quality": 0.3,
            "tool_accuracy": 0.3,
            "coherence": 0.2,
            "heuristic": 0.2
        }
        
        overall_score = 0.0
        total_weight = 0.0
        for metric, weight in weights.items():
            if metric in scores:
                overall_score += scores[metric] * weight
                total_weight += weight
        
        if total_weight > 0:
            overall_score /= total_weight
        
        unique_suggestions = self._deduplicate_suggestions(all_suggestions)
        
        return {
            "scores": {
                "overall": round(overall_score, 3),
                "response_quality": round(scores.get("response_quality", 0.0), 3),
                "tool_accuracy": round(scores.get("tool_accuracy", 0.0), 3),
                "coherence": round(scores.get("coherence", 0.0), 3)
            },
            "tool_evaluation": tool_evaluation,
            "issues_detected": all_issues,
            "improvement_suggestions": unique_suggestions,
            "evaluator_version": "1.0.0"
        }
    
    def _deduplicate_suggestions(self, suggestions: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """Remove duplicate suggestions"""
        seen = set()
        unique = []
        
        for suggestion in suggestions:
            key = f"{suggestion.get('type')}:{suggestion.get('suggestion', '')[:50]}"
            if key not in seen:
                seen.add(key)
                unique.append(suggestion)
        
        unique.sort(key=lambda x: x.get("confidence", 0.0), reverse=True)
        return unique
