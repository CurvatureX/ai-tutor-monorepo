"""
Doubao API router configuration
"""

from fastapi import APIRouter
from app.api.doubao.endpoints import conversation

router = APIRouter()

# Include Doubao endpoint routers
router.include_router(conversation.router, tags=["doubao"])
