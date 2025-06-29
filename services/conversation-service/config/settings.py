from pydantic_settings import BaseSettings
from typing import Optional
import os


class Settings(BaseSettings):
    """Application settings loaded from environment variables"""

    # Server configuration
    port: int = 8000
    host: str = "0.0.0.0"
    debug: bool = False

    # Local AI Model configuration
    use_local_model: bool = False
    model_path: Optional[str] = None
    model_name: str = "local-tutor"

    # Authentication
    jwt_secret: str = "your-secret-key-change-in-production"

    # External services
    user_service_url: Optional[str] = None
    redis_url: Optional[str] = None
    database_url: Optional[str] = None

    # Logging
    log_level: str = "INFO"

    # CORS
    cors_origins: list = ["*"]

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"
        case_sensitive = False

    def __init__(self, **kwargs):
        super().__init__(**kwargs)

        # Set defaults from environment
        self.use_local_model = os.getenv("USE_LOCAL_MODEL", "false").lower() == "true"
        if not self.model_path:
            self.model_path = os.getenv("MODEL_PATH", None)

        if not self.user_service_url:
            self.user_service_url = os.getenv(
                "USER_SERVICE_URL", "http://user-service:8001"
            )

        if not self.redis_url:
            self.redis_url = os.getenv("REDIS_URL", "redis://localhost:6379")

                self.debug = os.getenv("DEBUG", "false").lower() == "true"
