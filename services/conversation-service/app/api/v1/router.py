"""
V1 API router configuration
"""

from fastapi import APIRouter
from app.api.v1.endpoints import health, models, chat, language

router = APIRouter()

# Include endpoint routers
router.include_router(health.router, tags=["health"])
router.include_router(models.router, prefix="/v1", tags=["models"])
router.include_router(chat.router, prefix="/v1", tags=["chat"])
router.include_router(language.router, prefix="/v1", tags=["language"])
