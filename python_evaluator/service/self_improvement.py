"""Self-improvement service for generating suggestions"""
from typing import Dict, Any, List
from datetime import datetime, timedelta
from collections import defaultdict
import uuid


class SelfImprovementService:
    """Automatically generate improvement suggestions for prompts and tools"""
    
    async def analyze_and_generate_suggestions(
        self,
        lookback_days: int = 7
    ) -> Dict[str, Any]:
        """
        Analyze recent evaluations and generate improvement suggestions
        
        This is a simplified version - in production, this would connect
        to the database through the Go API
        """
        # Mock pattern detection (would fetch from database in production)
        patterns = self._detect_failure_patterns(lookback_days)
        
        prompt_suggestions = self._generate_prompt_suggestions(patterns)
        tool_suggestions = self._generate_tool_suggestions(patterns)
        
        all_suggestions = prompt_suggestions + tool_suggestions
        
        return {
            "status": "success",
            "analysis_period_days": lookback_days,
            "patterns_detected": len(patterns),
            "suggestions_generated": len(all_suggestions),
            "prompt_suggestions": len(prompt_suggestions),
            "tool_suggestions": len(tool_suggestions),
            "patterns": patterns,
            "suggestions": all_suggestions
        }
    
    def _detect_failure_patterns(self, lookback_days: int) -> List[Dict[str, Any]]:
        """Detect failure patterns (mock implementation)"""
        # In production, this would analyze actual evaluation data
        return [
            {
                "pattern_id": f"pattern_context_loss_{datetime.now().strftime('%Y%m')}",
                "type": "context_loss",
                "count": 15,
                "severity": "warning",
                "examples": []
            },
            {
                "pattern_id": f"pattern_hallucinated_params_{datetime.now().strftime('%Y%m')}",
                "type": "hallucinated_parameters",
                "count": 8,
                "severity": "warning",
                "examples": []
            }
        ]
    
    def _generate_prompt_suggestions(self, patterns: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """Generate prompt improvement suggestions"""
        suggestions = []
        
        suggestion_map = {
            "context_loss": {
                "suggestion": "Add explicit instruction: 'Maintain context from previous turns and reference user preferences throughout the conversation.'",
                "rationale": "Detected context loss issues. Explicit context maintenance instruction can help.",
                "confidence": 0.85,
                "expected_impact": "Reduce context_loss by 30-50%"
            },
            "ungrounded_parameters": {
                "suggestion": "Add parameter validation instruction: 'Only extract parameters explicitly mentioned by the user.'",
                "rationale": "Found ungrounded parameters. Stricter extraction rules needed.",
                "confidence": 0.80,
                "expected_impact": "Reduce parameter errors by 40%"
            },
            "hallucinated_parameters": {
                "suggestion": "Add schema awareness: 'Only use parameters defined in the tool schema.'",
                "rationale": "Detected hallucinated parameters. Explicit schema listing prevents this.",
                "confidence": 0.90,
                "expected_impact": "Reduce hallucinations by 50%"
            }
        }
        
        for pattern in patterns:
            pattern_type = pattern["type"]
            if pattern_type in suggestion_map:
                sugg_data = suggestion_map[pattern_type]
                suggestions.append({
                    "suggestion_id": f"sugg_{uuid.uuid4().hex[:12]}",
                    "type": "prompt",
                    "suggestion": sugg_data["suggestion"],
                    "rationale": f"{pattern['count']} occurrences detected. {sugg_data['rationale']}",
                    "confidence": sugg_data["confidence"],
                    "expected_impact": sugg_data["expected_impact"],
                    "pattern_detected": pattern
                })
        
        return suggestions
    
    def _generate_tool_suggestions(self, patterns: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """Generate tool schema suggestions"""
        suggestions = []
        
        tool_suggestion_map = {
            "missing_required_parameters": {
                "suggestion": "Review tool schema - mark frequently missing params as optional or improve descriptions",
                "rationale": "Missing required parameters detected. Schema clarity may need improvement.",
                "confidence": 0.80
            },
            "hallucinated_parameters": {
                "suggestion": "Add clear documentation to tool schema with supported parameters list",
                "rationale": "Hallucinated parameters detected. Schema needs better documentation.",
                "confidence": 0.85
            }
        }
        
        for pattern in patterns:
            pattern_type = pattern["type"]
            if pattern_type in tool_suggestion_map:
                sugg_data = tool_suggestion_map[pattern_type]
                suggestions.append({
                    "suggestion_id": f"sugg_{uuid.uuid4().hex[:12]}",
                    "type": "tool_schema",
                    "suggestion": sugg_data["suggestion"],
                    "rationale": sugg_data["rationale"],
                    "confidence": sugg_data["confidence"],
                    "pattern_detected": pattern
                })
        
        return suggestions
