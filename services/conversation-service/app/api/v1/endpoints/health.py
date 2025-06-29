"""
Health check endpoints
"""

from fastapi import APIRouter
from datetime import datetime
from app.api.deps import AIServiceDep, SettingsDep
from services.ai_service import AIService
from config.settings import Settings

router = APIRouter()


@router.get("/")
async def root():
    """Health check endpoint"""
    return {
        "service": "AI English Speaking Learning Service",
        "version": "1.0.0",
        "status": "healthy",
        "purpose": "English conversation practice and learning",
        "apis": ["OpenAI-compatible", "Doubao"],
        "backend": "External LLM APIs",
        "features": [
            "Pronunciation help",
            "Grammar practice",
            "Fluency building",
            "Natural conversation",
        ],
        "timestamp": datetime.utcnow().isoformat(),
    }


@router.get("/health")
async def health_check(
    ai_service: AIService = AIServiceDep, settings: Settings = SettingsDep
):
    """Detailed health check"""
    return {
        "status": "healthy",
        "services": {
            "ai_service": await ai_service.health_check(),
            "auth_service": True,  # TODO: Add auth service health check
            "doubao_service": True,
        },
        "config": {
            "default_model": settings.default_model,
            "max_tokens": settings.max_tokens,
            "temperature": settings.temperature,
        },
        "external_apis": {
            "doubao": settings.has_doubao_api(),
            "deepseek": settings.has_deepseek_api(),
            "gemini": settings.has_gemini_api(),
        },
        "timestamp": datetime.utcnow().isoformat(),
    }
