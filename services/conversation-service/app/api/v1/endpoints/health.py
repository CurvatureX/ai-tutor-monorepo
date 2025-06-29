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
        "backend": "Local AI Models + Doubao",
        "features": [
            "Pronunciation help",
            "Grammar practice", 
            "Fluency building",
            "Natural conversation"
        ],
        "timestamp": datetime.utcnow().isoformat(),
    }


@router.get("/health")
async def health_check(
    ai_service: AIService = AIServiceDep,
    settings: Settings = SettingsDep
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
            "use_local_model": settings.use_local_model,
            "model_name": settings.model_name,
            "doubao_available": True,
        },
        "timestamp": datetime.utcnow().isoformat(),
    }
