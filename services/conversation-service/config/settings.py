from pydantic_settings import BaseSettings
from typing import Optional, List
import os


class Settings(BaseSettings):
    """Application settings for AI English Speaking Learning Service"""

    # Server configuration
    port: int = 8000
    host: str = "0.0.0.0"
    debug: bool = False

    # Authentication
    jwt_secret: str = "your-secret-key-change-in-production"

    # External LLM API configurations
    doubao_api_key: Optional[str] = None
    doubao_base_url: str = "https://ark.cn-beijing.volces.com/api/v3"

    # DeepSeek API
    deepseek_api_key: Optional[str] = None
    deepseek_base_url: str = "https://api.deepseek.com/v1"

    # Gemini API (Google)
    gemini_api_key: Optional[str] = None
    gemini_base_url: str = "https://generativelanguage.googleapis.com/v1"

    # Default model configurations
    default_model: str = "doubao-seed-1-6-250615"
    max_tokens: int = 2000
    temperature: float = 0.7

    # External services
    user_service_url: Optional[str] = None
    redis_url: Optional[str] = None
    database_url: Optional[str] = None

    # Logging
    log_level: str = "INFO"

    # CORS
    cors_origins: List[str] = ["*"]

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"
        case_sensitive = False

    def __init__(self, **kwargs):
        super().__init__(**kwargs)

        # Load API keys from environment
        if not self.doubao_api_key:
            self.doubao_api_key = os.getenv("DOUBAO_API_KEY")

        if not self.deepseek_api_key:
            self.deepseek_api_key = os.getenv("DEEPSEEK_API_KEY")

        if not self.gemini_api_key:
            self.gemini_api_key = os.getenv("GEMINI_API_KEY")

        # Load service URLs from environment
        if not self.user_service_url:
            self.user_service_url = os.getenv(
                "USER_SERVICE_URL", "http://user-service:8001"
            )

        if not self.redis_url:
            self.redis_url = os.getenv("REDIS_URL", "redis://localhost:6379")

        # Debug mode
        self.debug = os.getenv("DEBUG", "false").lower() == "true"

    def has_doubao_api(self) -> bool:
        """Check if Doubao API is configured"""
        return self.doubao_api_key is not None and self.doubao_api_key.strip() != ""

    def has_deepseek_api(self) -> bool:
        """Check if DeepSeek API is configured"""
        return self.deepseek_api_key is not None and self.deepseek_api_key.strip() != ""

    def has_gemini_api(self) -> bool:
        """Check if Gemini API is configured"""
        return self.gemini_api_key is not None and self.gemini_api_key.strip() != ""
