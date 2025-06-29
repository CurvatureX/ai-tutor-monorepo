from fastapi import FastAPI, HTTPException, Depends
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from fastapi.responses import StreamingResponse
from pydantic import BaseModel
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
from doubao_client import DoubaoClient

# Initialize settings
settings = Settings()

# Initialize FastAPI app
app = FastAPI(
    title="AI Tutor Conversation Service",
    description="Unified API for AI tutoring conversations - supports both OpenAI-compatible API and Doubao API",
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
doubao_client = DoubaoClient()


# Doubao API models
class ConversationRequest(BaseModel):
    """对话请求模型"""

    message: str
    user_id: str
    model: Optional[str] = "doubao-seed-1-6-250615"


class MultimodalConversationRequest(BaseModel):
    """多模态对话请求模型"""

    text: str
    image_url: str
    user_id: str
    model: Optional[str] = "doubao-seed-1-6-250615"


@app.on_event("startup")
async def startup_event():
    """Initialize services on startup"""
    await ai_service.initialize()
    print(f"🚀 Conversation Service started on port {settings.port}")
    print(f"📚 Using local AI model: {settings.use_local_model}")
    print(f"🎯 Supporting both OpenAI-compatible and Doubao APIs")


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
        "apis": ["OpenAI-compatible", "Doubao"],
        "backend": "Local AI Models + Doubao",
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
            "doubao_service": True,
        },
        "config": {
            "use_local_model": settings.use_local_model,
            "model_name": settings.model_name,
            "doubao_available": True,
        },
        "timestamp": datetime.utcnow().isoformat(),
    }


# =============================================================================
# OpenAI-Compatible API Endpoints
# =============================================================================


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
        Model(
            id="doubao-seed-1-6-250615",
            object="model",
            created=int(datetime.utcnow().timestamp()),
            owned_by="doubao",
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
        # Check if this is a Doubao model request
        if request.model == "doubao-seed-1-6-250615":
            # Use Doubao API for this model
            user_message = request.messages[-1].content if request.messages else ""
            doubao_response_text = doubao_client.simple_chat(
                user_message, request.model
            )
            doubao_response = {
                "content": doubao_response_text,
                "usage": {
                    "prompt_tokens": 0,
                    "completion_tokens": 0,
                    "total_tokens": 0,
                },
            }

            response = ChatCompletionResponse(
                id=f"chatcmpl-{str(uuid.uuid4())}",
                object="chat.completion",
                created=int(datetime.utcnow().timestamp()),
                model=request.model,
                choices=[
                    ChatCompletionChoice(
                        index=0,
                        message=ChatMessage(
                            role="assistant", content=doubao_response["content"]
                        ),
                        finish_reason="stop",
                    )
                ],
                usage=Usage(
                    prompt_tokens=doubao_response.get("usage", {}).get(
                        "prompt_tokens", 0
                    ),
                    completion_tokens=doubao_response.get("usage", {}).get(
                        "completion_tokens", 0
                    ),
                    total_tokens=doubao_response.get("usage", {}).get(
                        "total_tokens", 0
                    ),
                ),
            )
            return response

        # Use local AI service for other models
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


# =============================================================================
# Doubao API Endpoints
# =============================================================================


@app.post("/conversation")
async def create_conversation(request: ConversationRequest):
    """
    创建对话 - Doubao API endpoint

    Args:
        request: 包含消息、用户ID和模型的请求体

    Returns:
        包含AI回复的响应
    """
    try:
        print(f"📝 Processing conversation request from user {request.user_id}")

        # 调用Doubao客户端
        response_text = doubao_client.simple_chat(
            request.message, request.model or "doubao-seed-1-6-250615"
        )
        response = {"content": response_text}

        print(f"✅ Got response: {response['content'][:100]}...")

        return {
            "status": "success",
            "response": response["content"],
            "user_id": request.user_id,
            "model": request.model,
            "timestamp": datetime.utcnow().isoformat(),
        }

    except Exception as e:
        print(f"❌ Error in conversation: {str(e)}")
        raise HTTPException(status_code=500, detail=f"对话处理失败: {str(e)}")


@app.post("/conversation/multimodal")
async def create_multimodal_conversation(request: MultimodalConversationRequest):
    """
    创建多模态对话 - Doubao API endpoint

    Args:
        request: 包含文本、图片URL、用户ID和模型的请求体

    Returns:
        包含AI回复的响应
    """
    try:
        print(f"🖼️ Processing multimodal conversation from user {request.user_id}")

        # 调用Doubao多模态客户端
        response = await doubao_client.chat_multimodal(
            request.text, request.image_url, request.model
        )

        print(f"✅ Got multimodal response: {response['content'][:100]}...")

        return {
            "status": "success",
            "response": response["content"],
            "user_id": request.user_id,
            "model": request.model,
            "timestamp": datetime.utcnow().isoformat(),
        }

    except Exception as e:
        print(f"❌ Error in multimodal conversation: {str(e)}")
        raise HTTPException(status_code=500, detail=f"多模态对话处理失败: {str(e)}")


@app.post("/test/doubao")
async def test_doubao_api():
    """
    测试Doubao API连接

    Returns:
        测试结果
    """
    try:
        print("🧪 Testing Doubao API connection...")

        test_message = "你好，这是一个API连接测试。"
        response_text = doubao_client.simple_chat(test_message)
        response = {"content": response_text}

        return {
            "status": "success",
            "message": "Doubao API连接测试成功!",
            "test_response": response["content"],
            "timestamp": datetime.utcnow().isoformat(),
        }

    except Exception as e:
        print(f"❌ Doubao API test failed: {str(e)}")
        return {
            "status": "error",
            "message": f"Doubao API连接测试失败: {str(e)}",
            "timestamp": datetime.utcnow().isoformat(),
        }


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.port,
        reload=settings.debug,
        log_level="info",
    )
