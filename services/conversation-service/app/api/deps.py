"""
API Dependencies - shared dependencies for API endpoints
"""

from fastapi import Depends
from app.core.security import get_current_user
from services.ai_service import AIService
from app.clients.doubao import DoubaoClient
from config.settings import Settings

# Global services (will be initialized in main.py)
ai_service: AIService = None
doubao_client: DoubaoClient = None
settings: Settings = None


def init_dependencies(app_settings: Settings, app_ai_service: AIService, app_doubao_client: DoubaoClient):
    """Initialize global dependencies"""
    global ai_service, doubao_client, settings
    settings = app_settings
    ai_service = app_ai_service
    doubao_client = app_doubao_client


def get_ai_service() -> AIService:
    """Get AI service instance"""
    return ai_service


def get_doubao_client() -> DoubaoClient:
    """Get Doubao client instance"""
    return doubao_client


def get_settings() -> Settings:
    """Get settings instance"""
    return settings


# Common dependencies
CurrentUser = Depends(get_current_user)
AIServiceDep = Depends(get_ai_service)
DoubaoClientDep = Depends(get_doubao_client)
SettingsDep = Depends(get_settings)
