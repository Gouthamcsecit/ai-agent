"""Configuration for Python Evaluator Service"""
import os
from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    """Application settings"""
    
    # Service
    service_host: str = "0.0.0.0"
    service_port: int = 8081
    debug: bool = False
    
    # Database
    database_url: str = "postgresql://postgres:postgres@localhost:5432/ai_agent_eval"
    
    # Redis
    redis_url: str = "redis://localhost:6379/0"
    
    # LLM
    openai_api_key: str = ""
    anthropic_api_key: str = ""
    llm_provider: str = "openai"
    llm_model: str = "gpt-4-turbo-preview"
    
    # Thresholds
    latency_threshold_ms: int = 1000
    min_quality_score: float = 0.7
    annotator_agreement_threshold: float = 0.8
    
    # Meta-Evaluation
    meta_eval_enabled: bool = True
    calibration_sample_size: int = 100
    
    class Config:
        env_file = ".env"
        case_sensitive = False


@lru_cache()
def get_settings() -> Settings:
    return Settings()
