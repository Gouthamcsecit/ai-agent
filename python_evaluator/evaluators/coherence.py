"""Multi-turn coherence evaluator"""
from typing import Dict, Any, List
from collections import Counter
from python_evaluator.evaluators.base import BaseEvaluator, EvaluationResult


class CoherenceEvaluator(BaseEvaluator):
    """
    Evaluates multi-turn conversation coherence:
    - Context maintenance across turns
    - Consistency (no contradictions)
    - Proper reference handling
    - Information retention
    """
    
    @property
    def evaluator_type(self) -> str:
        return "coherence"
    
    async def evaluate(self, conversation: Dict[str, Any]) -> EvaluationResult:
        """Evaluate conversation coherence"""
        turns = self.extract_turns(conversation)
        
        if len(turns) < 3:
            return EvaluationResult(
                evaluator_type=self.evaluator_type,
                score=1.0,
                details={"turns": len(turns), "note": "Too few turns for coherence evaluation"},
                confidence=0.5
            )
        
        issues = []
        suggestions = []
        
        context_score = self._evaluate_context_maintenance(turns, issues, suggestions)
        consistency_score = self._evaluate_consistency(turns, issues, suggestions)
        reference_score = self._evaluate_references(turns, issues, suggestions)
        retention_score = self._evaluate_information_retention(turns, issues, suggestions)
        
        overall_score = (
            context_score * 0.3 +
            consistency_score * 0.3 +
            reference_score * 0.2 +
            retention_score * 0.2
        )
        
        return EvaluationResult(
            evaluator_type=self.evaluator_type,
            score=overall_score,
            details={
                "context_maintenance": context_score,
                "consistency": consistency_score,
                "reference_handling": reference_score,
                "information_retention": retention_score,
                "total_turns": len(turns)
            },
            issues=issues,
            suggestions=suggestions,
            confidence=0.85
        )
    
    def _evaluate_context_maintenance(
        self,
        turns: List[Dict[str, Any]],
        issues: List[Dict[str, Any]],
        suggestions: List[Dict[str, Any]]
    ) -> float:
        """Evaluate if context is maintained across turns"""
        score = 1.0
        early_entities = self._extract_entities(turns[:min(3, len(turns))])
        
        if len(turns) > 5:
            later_content = " ".join([
                turn.get("content", "").lower()
                for turn in turns[5:]
                if turn.get("role") == "assistant"
            ])
            
            referenced_count = sum(
                1 for entity in early_entities
                if entity.lower() in later_content
            )
            
            if early_entities and referenced_count == 0:
                score = 0.6
                issues.append({
                    "type": "context_loss",
                    "severity": "warning",
                    "description": "Agent may have forgotten early context after turn 5",
                    "turn_id": 6
                })
                suggestions.append({
                    "type": "prompt",
                    "suggestion": "Add explicit instruction to maintain context from earlier in conversation",
                    "rationale": "Prevent context loss in long conversations",
                    "confidence": 0.75
                })
        
        return score
    
    def _evaluate_consistency(
        self,
        turns: List[Dict[str, Any]],
        issues: List[Dict[str, Any]],
        suggestions: List[Dict[str, Any]]
    ) -> float:
        """Check for contradictions in agent responses"""
        assistant_turns = [
            (idx, turn) for idx, turn in enumerate(turns)
            if turn.get("role") == "assistant"
        ]
        
        contradiction_keywords = [
            ("yes", "no"),
            ("can", "cannot"),
            ("will", "won't"),
            ("available", "unavailable"),
            ("possible", "impossible")
        ]
        
        score = 1.0
        for i, (idx1, turn1) in enumerate(assistant_turns):
            content1 = turn1.get("content", "").lower()
            for idx2, turn2 in assistant_turns[i+1:]:
                content2 = turn2.get("content", "").lower()
                
                for word1, word2 in contradiction_keywords:
                    if word1 in content1 and word2 in content2:
                        score = min(score, 0.7)
                        issues.append({
                            "type": "potential_contradiction",
                            "severity": "warning",
                            "description": f"Potential contradiction between turns {idx1+1} and {idx2+1}",
                            "turn_id": idx2 + 1
                        })
                        break
        
        return score
    
    def _evaluate_references(
        self,
        turns: List[Dict[str, Any]],
        issues: List[Dict[str, Any]],
        suggestions: List[Dict[str, Any]]
    ) -> float:
        """Evaluate handling of references and pronouns"""
        score = 1.0
        reference_words = ["it", "that", "this", "they", "them", "those", "these"]
        
        for idx, turn in enumerate(turns):
            if turn.get("role") == "assistant" and idx > 0:
                content = turn.get("content", "").lower()
                words = content.split()
                
                has_references = any(word in words for word in reference_words)
                
                if has_references:
                    prev_turn = turns[idx - 1]
                    if not prev_turn.get("content", "").strip():
                        score = min(score, 0.8)
                        issues.append({
                            "type": "unclear_reference",
                            "severity": "info",
                            "description": f"Turn {idx+1} uses references without clear antecedent",
                            "turn_id": idx + 1
                        })
        
        return score
    
    def _evaluate_information_retention(
        self,
        turns: List[Dict[str, Any]],
        issues: List[Dict[str, Any]],
        suggestions: List[Dict[str, Any]]
    ) -> float:
        """Evaluate if important information is retained"""
        score = 1.0
        user_provided_info = {}
        
        for idx, turn in enumerate(turns):
            if turn.get("role") == "user":
                content = turn.get("content", "")
                if any(word in content.lower() for word in ["my name is", "i'm", "prefer", "need"]):
                    user_provided_info[idx] = content
        
        if user_provided_info:
            for info_turn_idx, info in user_provided_info.items():
                acknowledged = False
                for turn in turns[info_turn_idx+1:info_turn_idx+4]:
                    if turn.get("role") == "assistant":
                        acknowledged = True
                        break
                
                if not acknowledged and len(turns) > info_turn_idx + 3:
                    score = min(score, 0.75)
                    issues.append({
                        "type": "information_not_retained",
                        "severity": "warning",
                        "description": f"User-provided information from turn {info_turn_idx+1} may not have been retained",
                        "turn_id": info_turn_idx + 1
                    })
        
        return score
    
    def _extract_entities(self, turns: List[Dict[str, Any]]) -> List[str]:
        """Extract key entities from turns"""
        entities = set()
        
        for turn in turns:
            content = turn.get("content", "")
            words = content.split()
            
            for word in words:
                if word and word[0].isupper() and len(word) > 2:
                    entities.add(word)
        
        return list(entities)
