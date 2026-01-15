"""Tool call evaluator for verifying tool usage accuracy"""
from typing import Dict, Any, List, Set
from python_evaluator.evaluators.base import BaseEvaluator, EvaluationResult


class ToolCallEvaluator(BaseEvaluator):
    """
    Evaluates tool calling behavior:
    - Correct tool selection
    - Parameter accuracy
    - Hallucinated parameters
    - Execution success
    """
    
    TOOL_SCHEMAS = {
        "flight_search": {
            "required": ["destination"],
            "optional": ["date_range", "departure_city", "passengers", "cabin_class"],
            "parameter_types": {
                "destination": "string",
                "date_range": "string",
                "departure_city": "string",
                "passengers": "integer",
                "cabin_class": "string"
            }
        },
        "hotel_search": {
            "required": ["location", "check_in", "check_out"],
            "optional": ["guests", "room_type", "max_price"],
            "parameter_types": {
                "location": "string",
                "check_in": "date",
                "check_out": "date",
                "guests": "integer",
                "room_type": "string",
                "max_price": "number"
            }
        },
        "booking_create": {
            "required": ["service_type", "item_id", "customer_info"],
            "optional": ["special_requests", "payment_method"],
            "parameter_types": {
                "service_type": "string",
                "item_id": "string",
                "customer_info": "object",
                "special_requests": "string",
                "payment_method": "string"
            }
        }
    }
    
    @property
    def evaluator_type(self) -> str:
        return "tool_call"
    
    async def evaluate(self, conversation: Dict[str, Any]) -> EvaluationResult:
        """Evaluate tool calls in conversation"""
        turns = self.extract_turns(conversation)
        all_tool_calls = self._extract_all_tool_calls(turns)
        
        if not all_tool_calls:
            return EvaluationResult(
                evaluator_type=self.evaluator_type,
                score=1.0,
                details={"tool_calls_count": 0},
                confidence=1.0
            )
        
        evaluations = []
        issues = []
        suggestions = []
        
        for turn_idx, tool_call in all_tool_calls:
            eval_result = self._evaluate_tool_call(tool_call, turns, turn_idx)
            evaluations.append(eval_result)
            issues.extend(eval_result.get("issues", []))
            suggestions.extend(eval_result.get("suggestions", []))
        
        selection_scores = [e["selection_accuracy"] for e in evaluations]
        parameter_scores = [e["parameter_accuracy"] for e in evaluations]
        execution_successes = [e["execution_success"] for e in evaluations]
        
        selection_accuracy = sum(selection_scores) / len(selection_scores)
        parameter_accuracy = sum(parameter_scores) / len(parameter_scores)
        execution_success_rate = sum(execution_successes) / len(execution_successes)
        
        overall_score = (
            selection_accuracy * 0.4 +
            parameter_accuracy * 0.4 +
            execution_success_rate * 0.2
        )
        
        return EvaluationResult(
            evaluator_type=self.evaluator_type,
            score=overall_score,
            details={
                "selection_accuracy": selection_accuracy,
                "parameter_accuracy": parameter_accuracy,
                "execution_success_rate": execution_success_rate,
                "total_tool_calls": len(evaluations),
                "hallucinated_parameters_count": sum(
                    len(e.get("hallucinated_params", [])) for e in evaluations
                )
            },
            issues=issues,
            suggestions=suggestions,
            confidence=0.9
        )
    
    def _extract_all_tool_calls(self, turns: List[Dict[str, Any]]) -> List[tuple]:
        """Extract all tool calls with their turn index"""
        tool_calls = []
        for idx, turn in enumerate(turns):
            if turn.get("role") == "assistant":
                for tc in turn.get("tool_calls", []):
                    tool_calls.append((idx, tc))
        return tool_calls
    
    def _evaluate_tool_call(
        self,
        tool_call: Dict[str, Any],
        turns: List[Dict[str, Any]],
        turn_idx: int
    ) -> Dict[str, Any]:
        """Evaluate a single tool call"""
        tool_name = tool_call.get("tool_name", "")
        parameters = tool_call.get("parameters", {})
        result = tool_call.get("result", {})
        
        if tool_name not in self.TOOL_SCHEMAS:
            return {
                "selection_accuracy": 0.5,
                "parameter_accuracy": 0.5,
                "execution_success": result.get("status") == "success",
                "issues": [{
                    "type": "unknown_tool",
                    "severity": "warning",
                    "description": f"Tool '{tool_name}' not in known schemas",
                    "turn_id": turn_idx + 1
                }],
                "suggestions": []
            }
        
        schema = self.TOOL_SCHEMAS[tool_name]
        issues = []
        suggestions = []
        
        selection_accuracy = 1.0
        param_eval = self._evaluate_parameters(parameters, schema, turns[:turn_idx+1], turn_idx)
        
        issues.extend(param_eval["issues"])
        suggestions.extend(param_eval["suggestions"])
        
        execution_success = result.get("status") == "success"
        if not execution_success:
            issues.append({
                "type": "tool_execution_failure",
                "severity": "critical",
                "description": f"Tool '{tool_name}' execution failed",
                "turn_id": turn_idx + 1
            })
        
        return {
            "selection_accuracy": selection_accuracy,
            "parameter_accuracy": param_eval["accuracy"],
            "execution_success": execution_success,
            "hallucinated_params": param_eval["hallucinated"],
            "issues": issues,
            "suggestions": suggestions
        }
    
    def _evaluate_parameters(
        self,
        parameters: Dict[str, Any],
        schema: Dict[str, Any],
        context_turns: List[Dict[str, Any]],
        turn_idx: int
    ) -> Dict[str, Any]:
        """Evaluate parameter accuracy"""
        issues = []
        suggestions = []
        hallucinated = []
        
        required_params = set(schema.get("required", []))
        optional_params = set(schema.get("optional", []))
        all_valid_params = required_params | optional_params
        provided_params = set(parameters.keys())
        
        missing_required = required_params - provided_params
        if missing_required:
            issues.append({
                "type": "missing_required_parameters",
                "severity": "critical",
                "description": f"Missing required parameters: {', '.join(missing_required)}",
                "turn_id": turn_idx + 1
            })
            suggestions.append({
                "type": "tool_schema",
                "suggestion": f"Ensure tool call includes required parameters: {', '.join(missing_required)}",
                "rationale": "Required parameters must be provided for successful execution",
                "confidence": 0.95
            })
        
        hallucinated_params = provided_params - all_valid_params
        if hallucinated_params:
            hallucinated.extend(list(hallucinated_params))
            issues.append({
                "type": "hallucinated_parameters",
                "severity": "warning",
                "description": f"Unknown parameters provided: {', '.join(hallucinated_params)}",
                "turn_id": turn_idx + 1
            })
            suggestions.append({
                "type": "prompt",
                "suggestion": "Add instruction to only use documented tool parameters",
                "rationale": "Prevent hallucinated parameters",
                "confidence": 0.85
            })
        
        context_check = self._check_parameter_context(parameters, context_turns)
        if not context_check["all_grounded"]:
            issues.append({
                "type": "ungrounded_parameters",
                "severity": "warning",
                "description": f"Parameters may not be grounded in context: {', '.join(context_check['ungrounded'])}",
                "turn_id": turn_idx + 1
            })
        
        total_expected = len(required_params) + len(optional_params)
        correct_params = len(provided_params & all_valid_params)
        missing_score = 1.0 - (len(missing_required) / max(len(required_params), 1))
        hallucination_penalty = len(hallucinated_params) * 0.1
        
        accuracy = max(0.0, (correct_params / max(total_expected, 1)) * missing_score - hallucination_penalty)
        
        return {
            "accuracy": min(1.0, accuracy),
            "hallucinated": hallucinated,
            "issues": issues,
            "suggestions": suggestions
        }
    
    def _check_parameter_context(
        self,
        parameters: Dict[str, Any],
        context_turns: List[Dict[str, Any]]
    ) -> Dict[str, Any]:
        """Check if parameters are grounded in conversation context"""
        user_content = " ".join([
            turn.get("content", "").lower()
            for turn in context_turns
            if turn.get("role") == "user"
        ])
        
        ungrounded = []
        for key, value in parameters.items():
            value_str = str(value).lower()
            if len(value_str) < 3:
                continue
            if value_str not in user_content:
                ungrounded.append(key)
        
        return {
            "all_grounded": len(ungrounded) == 0,
            "ungrounded": ungrounded
        }
