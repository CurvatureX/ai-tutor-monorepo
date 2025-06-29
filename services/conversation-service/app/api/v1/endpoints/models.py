"""
OpenAI-compatible models endpoint
"""

from fastapi import APIRouter
from datetime import datetime
from app.api.deps import CurrentUser
from models.openai_models import Model, ModelList

router = APIRouter()


@router.get("/models", response_model=ModelList)
async def list_models(user=CurrentUser):
    """List available AI models (OpenAI compatible)"""
    models = [
        Model(
            id="gpt-4-tutor",
            object="model",
            created=int(datetime.utcnow().timestamp()),
            owned_by="ai-tutor",
        ),
        Model(
            id="gpt-3.5-turbo-tutor",
            object="model",
            created=int(datetime.utcnow().timestamp()),
            owned_by="ai-tutor",
        ),
        Model(
            id="local-tutor",
            object="model",
            created=int(datetime.utcnow().timestamp()),
            owned_by="ai-tutor",
        ),
        Model(
            id="doubao-seed-1-6-250615",
            object="model",
            created=int(datetime.utcnow().timestamp()),
            owned_by="doubao",
        ),
    ]
    return ModelList(object="list", data=models)
