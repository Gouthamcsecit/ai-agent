#!/usr/bin/env python3
"""Generate sample conversation data for testing"""
import json
from datetime import datetime, timedelta
import random

def generate_sample_conversation(conv_id: str, agent_version: str = "v2.3.1"):
    """Generate a sample conversation"""
    
    scenarios = [
        {
            "user_messages": [
                "I need to book a flight to NYC next week",
                "Yes, departing from San Francisco",
                "Monday morning would be perfect"
            ],
            "assistant_responses": [
                "I'd be happy to help you book a flight to NYC. Could you provide your departure city?",
                "Great! Let me search for flights from San Francisco to NYC for next week.",
                "Perfect! I found several morning flights on Monday."
            ],
            "tools": [
                None,
                {
                    "tool_name": "flight_search",
                    "parameters": {"destination": "NYC", "departure_city": "San Francisco"},
                    "result": {"status": "success", "flights": ["AA123", "UA456"]},
                    "latency_ms": 450
                },
                None
            ],
            "rating": 4,
            "mission_completed": True
        },
        {
            "user_messages": [
                "Find me a hotel in Paris",
                "Check-in June 15, check-out June 20",
                "Actually, I meant July, not June"
            ],
            "assistant_responses": [
                "I'll help you find a hotel in Paris. When are your dates?",
                "Searching for hotels in Paris for June 15-20.",
                "No problem! Let me update the search for July 15-20."
            ],
            "tools": [
                None,
                {
                    "tool_name": "hotel_search",
                    "parameters": {"location": "Paris", "check_in": "2024-06-15", "check_out": "2024-06-20"},
                    "result": {"status": "success", "hotels": ["Hotel A"]},
                    "latency_ms": 380
                },
                {
                    "tool_name": "hotel_search",
                    "parameters": {"location": "Paris", "check_in": "2024-07-15", "check_out": "2024-07-20"},
                    "result": {"status": "success", "hotels": ["Hotel C"]},
                    "latency_ms": 420
                }
            ],
            "rating": 5,
            "mission_completed": True
        }
    ]
    
    scenario = random.choice(scenarios)
    
    turns = []
    turn_id = 1
    base_time = datetime.utcnow() - timedelta(hours=2)
    
    for i in range(len(scenario["user_messages"])):
        turns.append({
            "turn_id": turn_id,
            "role": "user",
            "content": scenario["user_messages"][i],
            "timestamp": (base_time + timedelta(seconds=turn_id * 30)).isoformat() + "Z"
        })
        turn_id += 1
        
        assistant_turn = {
            "turn_id": turn_id,
            "role": "assistant",
            "content": scenario["assistant_responses"][i],
            "timestamp": (base_time + timedelta(seconds=turn_id * 30)).isoformat() + "Z"
        }
        
        if scenario["tools"][i]:
            assistant_turn["tool_calls"] = [scenario["tools"][i]]
        
        turns.append(assistant_turn)
        turn_id += 1
    
    total_latency = sum(
        tc.get("latency_ms", 0)
        for turn in turns
        if turn.get("tool_calls")
        for tc in turn["tool_calls"]
    ) + random.randint(200, 500)
    
    return {
        "conversation_id": conv_id,
        "agent_version": agent_version,
        "turns": turns,
        "feedback": {
            "user_rating": scenario["rating"],
            "ops_review": {"quality": "good" if scenario["rating"] >= 4 else "needs_improvement"},
            "annotations": [{"type": "tool_accuracy", "label": "correct", "annotator_id": "ann_001"}]
        },
        "metadata": {
            "total_latency_ms": total_latency,
            "mission_completed": scenario["mission_completed"]
        }
    }


def generate_batch(count: int = 10):
    """Generate batch of sample conversations"""
    return [
        generate_sample_conversation(f"conv_sample_{i+1:03d}")
        for i in range(count)
    ]


if __name__ == "__main__":
    batch = generate_batch(20)
    
    with open("sample_conversations.json", "w") as f:
        json.dump(batch, f, indent=2)
    
    print(f"Generated {len(batch)} sample conversations")
    print("Saved to: sample_conversations.json")
