"""
OpenAI-compatible models endpoint
"""

from fastapi import APIRouter
from datetime import datetime
from app.api.deps import CurrentUser, SettingsDep
from models.openai_models import Model, ModelList
from config.settings import Settings

router = APIRouter()


@router.get("/models", response_model=ModelList)
async def list_models(user=CurrentUser, settings: Settings = SettingsDep):
    """List available AI models (OpenAI compatible)"""
    models = []

    # Add Doubao model if configured
    if settings.has_doubao_api():
        models.append(
            Model(
                id="doubao-seed-1-6-250615",
                object="model",
                created=int(datetime.utcnow().timestamp()),
                owned_by="doubao",
            )
        )

    # Add DeepSeek model if configured
    if settings.has_deepseek_api():
        models.append(
            Model(
                id="deepseek-chat",
                object="model",
                created=int(datetime.utcnow().timestamp()),
                owned_by="deepseek",
            )
        )

    # Add Gemini model if configured
    if settings.has_gemini_api():
        models.append(
            Model(
                id="gemini-2.5-flash",
                object="model",
                created=int(datetime.utcnow().timestamp()),
                owned_by="google",
            )
        )

    # Always add a local fallback model for demo purposes
    models.append(
        Model(
            id="local-tutor",
            object="model",
            created=int(datetime.utcnow().timestamp()),
            owned_by="ai-tutor",
        )
    )

    return ModelList(object="list", data=models)
