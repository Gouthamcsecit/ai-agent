"""Heuristic evaluator for format and performance checks"""
from typing import Dict, Any, List
from python_evaluator.evaluators.base import BaseEvaluator, EvaluationResult
from python_evaluator.config import get_settings

settings = get_settings()


class HeuristicEvaluator(BaseEvaluator):
    """
    Fast rule-based checks:
    - Format compliance
    - Latency thresholds
    - Required fields presence
    - Basic quality checks
    """
    
    @property
    def evaluator_type(self) -> str:
        return "heuristic"
    
    async def evaluate(self, conversation: Dict[str, Any]) -> EvaluationResult:
        """Evaluate using heuristic rules"""
        turns = self.extract_turns(conversation)
        metadata = self.get_metadata(conversation)
        
        issues = []
        suggestions = []
        checks_passed = []
        checks_failed = []
        
        latency_score = self._check_latency(metadata, issues, checks_passed, checks_failed)
        format_score = self._check_format_compliance(turns, issues, checks_passed, checks_failed)
        fields_score = self._check_required_fields(conversation, issues, checks_passed, checks_failed)
        length_score = self._check_response_lengths(turns, issues, checks_passed, checks_failed)
        tool_score = self._check_tool_call_format(turns, issues, checks_passed, checks_failed)
        
        overall_score = (
            latency_score * 0.2 +
            format_score * 0.2 +
            fields_score * 0.2 +
            length_score * 0.2 +
            tool_score * 0.2
        )
        
        return EvaluationResult(
            evaluator_type=self.evaluator_type,
            score=overall_score,
            details={
                "checks_passed": len(checks_passed),
                "checks_failed": len(checks_failed),
                "passed": checks_passed,
                "failed": checks_failed
            },
            issues=issues,
            suggestions=suggestions,
            confidence=1.0
        )
    
    def _check_latency(
        self,
        metadata: Dict[str, Any],
        issues: List[Dict[str, Any]],
        passed: List[str],
        failed: List[str]
    ) -> float:
        """Check if latency is within threshold"""
        total_latency = metadata.get("total_latency_ms", 0)
        threshold = get_settings().latency_threshold_ms
        
        if total_latency <= threshold:
            passed.append("latency_check")
            return 1.0
        else:
            failed.append("latency_check")
            issues.append({
                "type": "latency",
                "severity": "warning" if total_latency < threshold * 1.5 else "critical",
                "description": f"Response latency {total_latency}ms exceeds {threshold}ms target"
            })
            return max(0.0, 1.0 - (total_latency - threshold) / threshold)
    
    def _check_format_compliance(
        self,
        turns: List[Dict[str, Any]],
        issues: List[Dict[str, Any]],
        passed: List[str],
        failed: List[str]
    ) -> float:
        """Check if responses follow expected format"""
        score = 1.0
        
        for idx, turn in enumerate(turns):
            if turn.get("role") == "assistant":
                content = turn.get("content", "")
                
                if not content or not content.strip():
                    failed.append(f"empty_response_turn_{idx+1}")
                    issues.append({
                        "type": "format_violation",
                        "severity": "critical",
                        "description": f"Turn {idx+1} has empty response",
                        "turn_id": idx + 1
                    })
                    score = min(score, 0.5)
                    continue
                
                if len(content) > 10 and not content.rstrip()[-1] in ".!?":
                    failed.append(f"incomplete_sentence_turn_{idx+1}")
                    issues.append({
                        "type": "format_violation",
                        "severity": "info",
                        "description": f"Turn {idx+1} response may be incomplete",
                        "turn_id": idx + 1
                    })
                    score = min(score, 0.9)
                else:
                    passed.append(f"format_turn_{idx+1}")
        
        return score
    
    def _check_required_fields(
        self,
        conversation: Dict[str, Any],
        issues: List[Dict[str, Any]],
        passed: List[str],
        failed: List[str]
    ) -> float:
        """Check presence of required fields"""
        required_fields = {
            "conversation_id": conversation.get("conversation_id"),
            "agent_version": conversation.get("agent_version"),
            "turns": conversation.get("turns")
        }
        
        missing = [field for field, value in required_fields.items() if not value]
        
        if not missing:
            passed.append("required_fields")
            return 1.0
        else:
            failed.append("required_fields")
            issues.append({
                "type": "missing_fields",
                "severity": "critical",
                "description": f"Missing required fields: {', '.join(missing)}"
            })
            return max(0.0, 1.0 - (len(missing) * 0.3))
    
    def _check_response_lengths(
        self,
        turns: List[Dict[str, Any]],
        issues: List[Dict[str, Any]],
        passed: List[str],
        failed: List[str]
    ) -> float:
        """Check if response lengths are reasonable"""
        score = 1.0
        
        for idx, turn in enumerate(turns):
            if turn.get("role") == "assistant":
                content = turn.get("content", "")
                length = len(content)
                
                if length < 10 and length > 0:
                    failed.append(f"short_response_turn_{idx+1}")
                    issues.append({
                        "type": "response_quality",
                        "severity": "warning",
                        "description": f"Turn {idx+1} has very short response ({length} chars)",
                        "turn_id": idx + 1
                    })
                    score = min(score, 0.8)
                elif length > 2000:
                    failed.append(f"long_response_turn_{idx+1}")
                    issues.append({
                        "type": "response_quality",
                        "severity": "info",
                        "description": f"Turn {idx+1} has very long response ({length} chars)",
                        "turn_id": idx + 1
                    })
                    score = min(score, 0.9)
                else:
                    passed.append(f"length_turn_{idx+1}")
        
        return score
    
    def _check_tool_call_format(
        self,
        turns: List[Dict[str, Any]],
        issues: List[Dict[str, Any]],
        passed: List[str],
        failed: List[str]
    ) -> float:
        """Check tool call formatting"""
        score = 1.0
        
        for idx, turn in enumerate(turns):
            if turn.get("role") == "assistant":
                tool_calls = turn.get("tool_calls", [])
                
                for tc_idx, tool_call in enumerate(tool_calls):
                    if not tool_call.get("tool_name"):
                        failed.append(f"tool_name_missing_turn_{idx+1}")
                        issues.append({
                            "type": "tool_call_format",
                            "severity": "critical",
                            "description": f"Turn {idx+1} tool call missing tool_name",
                            "turn_id": idx + 1
                        })
                        score = min(score, 0.5)
                    
                    if "parameters" not in tool_call:
                        failed.append(f"parameters_missing_turn_{idx+1}")
                        issues.append({
                            "type": "tool_call_format",
                            "severity": "critical",
                            "description": f"Turn {idx+1} tool call missing parameters",
                            "turn_id": idx + 1
                        })
                        score = min(score, 0.5)
                    else:
                        passed.append(f"tool_format_turn_{idx+1}_call_{tc_idx}")
        
        return score
