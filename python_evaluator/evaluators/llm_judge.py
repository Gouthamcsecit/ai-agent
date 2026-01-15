"""LLM-as-Judge evaluator for response quality assessment"""
import json
import re
from typing import Dict, Any, List
from openai import AsyncOpenAI
from anthropic import AsyncAnthropic
from python_evaluator.evaluators.base import BaseEvaluator, EvaluationResult
from python_evaluator.config import get_settings

settings = get_settings()


class LLMJudgeEvaluator(BaseEvaluator):
    """
    Uses LLM to assess:
    - Response quality
    - Helpfulness
    - Factuality
    - Appropriateness
    """
    
    def __init__(self, config: Dict[str, Any] = None):
        super().__init__(config)
        self.settings = get_settings()
        
        if self.settings.llm_provider == "openai" and self.settings.openai_api_key:
            self.client = AsyncOpenAI(api_key=self.settings.openai_api_key)
            self.provider = "openai"
        elif self.settings.anthropic_api_key:
            self.client = AsyncAnthropic(api_key=self.settings.anthropic_api_key)
            self.provider = "anthropic"
        else:
            self.client = None
            self.provider = None
    
    @property
    def evaluator_type(self) -> str:
        return "llm_judge"
    
    async def evaluate(self, conversation: Dict[str, Any]) -> EvaluationResult:
        """Evaluate conversation using LLM as judge"""
        turns = self.extract_turns(conversation)
        
        if not self.client:
            # Return mock evaluation if no LLM configured
            return self._mock_evaluation()
        
        try:
            prompt = self._build_evaluation_prompt(turns)
            judgment = await self._get_llm_judgment(prompt)
            
            scores = judgment.get("scores", {})
            issues = judgment.get("issues", [])
            suggestions = judgment.get("suggestions", [])
            
            overall_score = (
                scores.get("helpfulness", 0.5) * 0.3 +
                scores.get("factuality", 0.5) * 0.3 +
                scores.get("clarity", 0.5) * 0.2 +
                scores.get("appropriateness", 0.5) * 0.2
            )
            
            return EvaluationResult(
                evaluator_type=self.evaluator_type,
                score=overall_score,
                details={
                    "helpfulness": scores.get("helpfulness"),
                    "factuality": scores.get("factuality"),
                    "clarity": scores.get("clarity"),
                    "appropriateness": scores.get("appropriateness"),
                    "reasoning": judgment.get("reasoning", "")
                },
                issues=issues,
                suggestions=suggestions,
                confidence=judgment.get("confidence", 0.8)
            )
        except Exception as e:
            return EvaluationResult(
                evaluator_type=self.evaluator_type,
                score=0.5,
                details={"error": str(e)},
                confidence=0.0
            )
    
    def _mock_evaluation(self) -> EvaluationResult:
        """Return mock evaluation when no LLM is configured"""
        import random
        score = random.uniform(0.7, 0.95)
        return EvaluationResult(
            evaluator_type=self.evaluator_type,
            score=score,
            details={
                "helpfulness": score + random.uniform(-0.1, 0.1),
                "factuality": score + random.uniform(-0.1, 0.1),
                "clarity": score + random.uniform(-0.1, 0.1),
                "appropriateness": score + random.uniform(-0.1, 0.1),
                "reasoning": "Mock evaluation - no LLM configured"
            },
            confidence=0.5
        )
    
    def _build_evaluation_prompt(self, turns: List[Dict[str, Any]]) -> str:
        """Build prompt for LLM judge"""
        conversation_text = self._format_conversation(turns)
        
        return f"""You are an expert evaluator assessing AI agent conversations. Evaluate the following conversation.

Conversation:
{conversation_text}

Evaluate the AI assistant's responses on these dimensions (score 0.0 to 1.0):
1. Helpfulness: Does it address the user's needs effectively?
2. Factuality: Are the statements accurate and truthful?
3. Clarity: Is the response clear and easy to understand?
4. Appropriateness: Is the tone and approach suitable?

Also identify:
- Any issues or problems in the responses
- Suggestions for improvement

Respond in JSON format:
{{
    "scores": {{
        "helpfulness": <0.0-1.0>,
        "factuality": <0.0-1.0>,
        "clarity": <0.0-1.0>,
        "appropriateness": <0.0-1.0>
    }},
    "reasoning": "<brief explanation>",
    "issues": [
        {{"type": "<issue_type>", "severity": "<critical|warning|info>", "description": "<description>"}}
    ],
    "suggestions": [
        {{"type": "prompt", "suggestion": "<suggestion>", "rationale": "<rationale>", "confidence": <0.0-1.0>}}
    ],
    "confidence": <0.0-1.0>
}}"""
    
    def _format_conversation(self, turns: List[Dict[str, Any]]) -> str:
        """Format conversation turns for prompt"""
        formatted = []
        for turn in turns:
            role = turn.get("role", "unknown")
            content = turn.get("content", "")
            tool_calls = turn.get("tool_calls", [])
            
            formatted.append(f"{role.upper()}: {content}")
            
            if tool_calls:
                for tc in tool_calls:
                    formatted.append(f"  [Tool: {tc.get('tool_name')} with params: {tc.get('parameters')}]")
        
        return "\n".join(formatted)
    
    async def _get_llm_judgment(self, prompt: str) -> Dict[str, Any]:
        """Get judgment from LLM"""
        if self.provider == "openai":
            response = await self.client.chat.completions.create(
                model=self.settings.llm_model,
                messages=[
                    {"role": "system", "content": "You are an expert AI conversation evaluator. Always respond in valid JSON format."},
                    {"role": "user", "content": prompt}
                ],
                temperature=0.3,
                response_format={"type": "json_object"}
            )
            content = response.choices[0].message.content
        elif self.provider == "anthropic":
            response = await self.client.messages.create(
                model=self.settings.llm_model,
                max_tokens=2000,
                temperature=0.3,
                messages=[{"role": "user", "content": prompt}]
            )
            content = response.content[0].text
        else:
            raise ValueError("No LLM provider configured")
        
        try:
            return json.loads(content)
        except json.JSONDecodeError:
            json_match = re.search(r'\{.*\}', content, re.DOTALL)
            if json_match:
                return json.loads(json_match.group())
            raise ValueError("Could not parse LLM response as JSON")
