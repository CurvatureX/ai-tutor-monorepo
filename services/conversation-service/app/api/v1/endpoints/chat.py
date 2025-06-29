"""
OpenAI-compatible chat completion endpoints
"""

from fastapi import APIRouter, HTTPException
from fastapi.responses import StreamingResponse
import json
import uuid
from datetime import datetime

from app.api.deps import CurrentUser, AIServiceDep, DoubaoClientDep
from models.openai_models import (
    ChatCompletionRequest,
    ChatCompletionResponse,
    ChatCompletionChoice,
    ChatMessage,
    Usage,
)
from services.ai_service import AIService
from app.clients.doubao import DoubaoClient

router = APIRouter()


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


@router.post("/chat/completions")
async def create_chat_completion(
    request: ChatCompletionRequest,
    user=CurrentUser,
    ai_service: AIService = AIServiceDep,
    doubao_client: DoubaoClient = DoubaoClientDep,
):
    """Create a chat completion (OpenAI compatible)"""
    try:
        user_message = request.messages[-1].content if request.messages else ""

        # Route to appropriate service based on model
        if request.model == "doubao-seed-1-6-250615":
            # Use Doubao API for this model
            doubao_response_text = doubao_client.simple_chat(
                user_message, request.model
            )
            response_content = doubao_response_text

        elif request.model == "deepseek-chat":
            # Use DeepSeek model (for now, using AI service with model info)
            completion = await ai_service.create_completion(
                messages=request.messages,
                model=request.model,
                user_id=user.get("user_id"),
                temperature=request.temperature,
                max_tokens=request.max_tokens,
                stream=request.stream,
            )

            # Handle streaming response for DeepSeek
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

            response_content = completion["content"]

        elif request.model == "gemini-2.5-flash":
            # Use Gemini model (for now, using AI service with model info)
            completion = await ai_service.create_completion(
                messages=request.messages,
                model=request.model,
                user_id=user.get("user_id"),
                temperature=request.temperature,
                max_tokens=request.max_tokens,
                stream=request.stream,
            )

            # Handle streaming response for Gemini
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

            response_content = completion["content"]

        else:
            # Use local AI service for other models (local-tutor, etc.)
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

            response_content = completion["content"]

        # Create standard response for all models
        response = ChatCompletionResponse(
            id=f"chatcmpl-{str(uuid.uuid4())}",
            object="chat.completion",
            created=int(datetime.utcnow().timestamp()),
            model=request.model,
            choices=[
                ChatCompletionChoice(
                    index=0,
                    message=ChatMessage(role="assistant", content=response_content),
                    finish_reason="stop",
                )
            ],
            usage=Usage(
                prompt_tokens=(
                    completion.get("prompt_tokens", 0)
                    if 'completion' in locals()
                    else 0
                ),
                completion_tokens=(
                    completion.get("completion_tokens", 0)
                    if 'completion' in locals()
                    else 0
                ),
                total_tokens=(
                    completion.get("total_tokens", 0) if 'completion' in locals() else 0
                ),
            ),
        )
        return response

    except Exception as e:
        raise HTTPException(
            status_code=500, detail=f"Error creating completion: {str(e)}"
        )
