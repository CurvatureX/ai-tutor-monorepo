from fastapi import FastAPI, HTTPException, Depends
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from fastapi.responses import StreamingResponse
import uvicorn
import os
import json
from datetime import datetime
import uuid
import asyncio
from typing import Optional

from models.openai_models import (
    ChatCompletionRequest,
    ChatCompletionResponse,
    ChatCompletionChoice,
    ChatMessage,
    Usage,
    Model,
    ModelList,
)
from services.ai_service import AIService
from services.auth_service import AuthService
from config.settings import Settings

# Initialize settings
settings = Settings()

# Initialize FastAPI app
app = FastAPI(
    title="AI Tutor Conversation Service",
    description="OpenAI-compatible API for AI tutoring conversations using local models",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure appropriately for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Security
security = HTTPBearer()

# Initialize services
ai_service = AIService(settings)
auth_service = AuthService(settings)


@app.on_event("startup")
async def startup_event():
    """Initialize services on startup"""
    await ai_service.initialize()
    print(f"🚀 Conversation Service started on port {settings.port}")
    print(f"📚 Using local AI model: {settings.use_local_model}")


@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on shutdown"""
    await ai_service.cleanup()


async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
):
    """Validate JWT token and return user info"""
    try:
        user = await auth_service.validate_token(credentials.credentials)
        return user
    except Exception as e:
        raise HTTPException(
            status_code=401, detail="Invalid authentication credentials"
        )


@app.get("/")
async def root():
    """Health check endpoint"""
    return {
        "service": "AI Tutor Conversation Service",
        "version": "1.0.0",
        "status": "healthy",
        "backend": "Local AI Models",
        "timestamp": datetime.utcnow().isoformat(),
    }


@app.get("/health")
async def health_check():
    """Detailed health check"""
    return {
        "status": "healthy",
        "services": {
            "ai_service": await ai_service.health_check(),
            "auth_service": await auth_service.health_check(),
        },
        "config": {
            "use_local_model": settings.use_local_model,
            "model_name": settings.model_name,
        },
        "timestamp": datetime.utcnow().isoformat(),
    }


@app.get("/v1/models", response_model=ModelList)
async def list_models(user=Depends(get_current_user)):
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
    ]
    return ModelList(object="list", data=models)


async def stream_chat_completion(stream_generator):
    """Stream chat completion chunks in OpenAI format"""
    try:
        async for chunk in stream_generator:
            chunk_json = json.dumps(chunk, ensure_ascii=False)
            yield f"data: {chunk_json}\n\n"
        yield "data: [DONE]\n\n"
    except Exception as e:
        error_chunk = {"error": {"message": str(e), "type": "internal_error"}}
        yield f"data: {json.dumps(error_chunk)}\n\n"


@app.post("/v1/chat/completions")
async def create_chat_completion(
    request: ChatCompletionRequest, user=Depends(get_current_user)
):
    """Create a chat completion (OpenAI compatible)"""
    try:
        # Generate completion using AI service
        completion = await ai_service.create_completion(
            messages=request.messages,
            model=request.model,
            user_id=user.get("user_id"),
            temperature=request.temperature,
            max_tokens=request.max_tokens,
            stream=request.stream,
        )

        # Handle streaming response
        if request.stream:
            stream_generator = completion.get("stream")
            if stream_generator:
                return StreamingResponse(
                    stream_chat_completion(stream_generator),
                    media_type="text/plain",
                    headers={
                        "Cache-Control": "no-cache",
                        "Connection": "keep-alive",
                        "Content-Type": "text/plain; charset=utf-8",
                    },
                )

        # Create standard response
        response = ChatCompletionResponse(
            id=f"chatcmpl-{str(uuid.uuid4())}",
            object="chat.completion",
            created=int(datetime.utcnow().timestamp()),
            model=request.model,
            choices=[
                ChatCompletionChoice(
                    index=0,
                    message=ChatMessage(
                        role="assistant", content=completion["content"]
                    ),
                    finish_reason="stop",
                )
            ],
            usage=Usage(
                prompt_tokens=completion.get("prompt_tokens", 0),
                completion_tokens=completion.get("completion_tokens", 0),
                total_tokens=completion.get("total_tokens", 0),
            ),
        )

        return response

    except Exception as e:
        raise HTTPException(
            status_code=500, detail=f"Error creating completion: {str(e)}"
        )


@app.post("/v1/tutor/session")
async def create_tutor_session(
    subject: str, level: str = "intermediate", user=Depends(get_current_user)
):
    """Create a new tutoring session"""
    try:
        session = await ai_service.create_tutor_session(
            user_id=user.get("user_id"), subject=subject, level=level
        )
        return session
    except Exception as e:
        raise HTTPException(
            status_code=500, detail=f"Error creating tutor session: {str(e)}"
        )


@app.get("/v1/tutor/sessions")
async def get_user_sessions(user=Depends(get_current_user)):
    """Get user's tutoring sessions"""
    try:
        sessions = await ai_service.get_user_sessions(user.get("user_id"))
        return {"sessions": sessions}
    except Exception as e:
        raise HTTPException(
            status_code=500, detail=f"Error fetching sessions: {str(e)}"
        )


@app.get("/v1/tutor/sessions/{session_id}")
async def get_session_history(session_id: str, user=Depends(get_current_user)):
    """Get conversation history for a session"""
    try:
        history = await ai_service.get_session_history(session_id, user.get("user_id"))
        return {"history": history}
    except Exception as e:
        raise HTTPException(
            status_code=500, detail=f"Error fetching session history: {str(e)}"
        )


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.port,
        reload=settings.debug,
        log_level="info",
    )
