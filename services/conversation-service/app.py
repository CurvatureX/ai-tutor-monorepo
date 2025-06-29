"""
AI English Speaking Learning Service - Simplified Application

A consolidated FastAPI application for AI-powered English conversation practice.
All functionality combined into a single file for simplicity.

Version: 2.0.0 (Simplified)
"""

import os
import json
import uuid
import time
import asyncio
import httpx
from datetime import datetime
from typing import List, Dict, Any, Optional, AsyncGenerator, Union
from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI, APIRouter, HTTPException, Depends, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import StreamingResponse
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel, Field
from enum import Enum

# Import our prompt management system
from prompt import (
    create_system_prompt,
    create_conversation_starter,
    create_error_correction_prompt,
    get_encouragement_phrases,
    ConversationContext,
    LearningLevel,
    BASE_ENGLISH_TUTOR_PROMPT,
)


# =====================================================
# MODELS - All Pydantic models in one place
# =====================================================


class Role(str, Enum):
    SYSTEM = "system"
    USER = "user"
    ASSISTANT = "assistant"


class ChatMessage(BaseModel):
    role: Role
    content: str
    name: Optional[str] = None


class ChatCompletionRequest(BaseModel):
    model: str = Field(default="gpt-4-tutor", description="AI model to use")
    messages: List[ChatMessage] = Field(
        ..., description="List of messages in the conversation"
    )
    temperature: Optional[float] = Field(default=0.7, ge=0.0, le=2.0)
    max_tokens: Optional[int] = Field(default=None, ge=1)
    stream: Optional[bool] = Field(default=False)
    user: Optional[str] = Field(default=None)


class ChatCompletionChoice(BaseModel):
    index: int
    message: ChatMessage
    finish_reason: Optional[str] = None


class Usage(BaseModel):
    prompt_tokens: int
    completion_tokens: int
    total_tokens: int


class ChatCompletionResponse(BaseModel):
    id: str
    object: str = "chat.completion"
    created: int
    model: str
    choices: List[ChatCompletionChoice]
    usage: Usage


class Model(BaseModel):
    id: str
    object: str = "model"
    created: int
    owned_by: str


class ModelList(BaseModel):
    object: str = "list"
    data: List[Model]


class LanguageSession(BaseModel):
    id: str
    user_id: str
    language: str = "English"
    level: str = "intermediate"
    context: str = "general"  # conversation context
    topic: Optional[str] = None  # specific topic
    goals: List[str] = []
    interests: List[str] = []  # user interests
    created_at: str
    last_activity: str
    message_count: int = 0
    status: str = "active"


# =====================================================
# CONFIGURATION - Simple configuration class
# =====================================================


class Config:
    """Simple configuration class"""

    def __init__(self):
        self.host = os.getenv("HOST", "0.0.0.0")
        self.port = int(os.getenv("PORT", 8000))
        self.debug = os.getenv("DEBUG", "false").lower() == "true"
        self.log_level = os.getenv("LOG_LEVEL", "info")
        self.temperature = float(os.getenv("TEMPERATURE", "0.7"))
        self.max_tokens = int(os.getenv("MAX_TOKENS", "2000"))

        # API Keys
        self.doubao_api_key = os.getenv("DOUBAO_API_KEY")
        self.deepseek_api_key = os.getenv("DEEPSEEK_API_KEY")
        self.gemini_api_key = os.getenv("GEMINI_API_KEY")
        self.jwt_secret = os.getenv("JWT_SECRET", "your-secret-key")

    def has_doubao_api(self) -> bool:
        return bool(self.doubao_api_key)

    def has_deepseek_api(self) -> bool:
        return bool(self.deepseek_api_key)

    def has_gemini_api(self) -> bool:
        return bool(self.gemini_api_key)


# =====================================================
# SERVICES - All business logic in one class
# =====================================================


class ConversationService:
    """Consolidated service for all conversation functionality"""

    def __init__(self, config: Config):
        self.config = config
        self.sessions_store = {}  # In production, use Redis or database
        self.conversation_history = {}

    async def initialize(self):
        """Initialize the service"""
        print("🔄 Initializing Conversation Service...")

        available_apis = []
        if self.config.has_doubao_api():
            available_apis.append("Doubao")
        if self.config.has_deepseek_api():
            available_apis.append("DeepSeek")
        if self.config.has_gemini_api():
            available_apis.append("Gemini")

        if available_apis:
            print(f"✅ Service initialized with APIs: {', '.join(available_apis)}")
        else:
            print("⚠️  No external API keys configured - using demo responses")

    async def cleanup(self):
        """Cleanup resources"""
        print("🧹 Cleaning up Conversation Service...")

    def get_available_models(self) -> List[Model]:
        """Get list of available models"""
        models = [
            Model(
                id="gpt-4-tutor",
                object="model",
                created=int(time.time()),
                owned_by="local",
            ),
            Model(
                id="local-tutor",
                object="model",
                created=int(time.time()),
                owned_by="local",
            ),
        ]

        if self.config.has_doubao_api():
            models.append(
                Model(
                    id="doubao-seed-1-6-250615",
                    object="model",
                    created=int(time.time()),
                    owned_by="doubao",
                )
            )
        if self.config.has_deepseek_api():
            models.append(
                Model(
                    id="deepseek-chat",
                    object="model",
                    created=int(time.time()),
                    owned_by="deepseek",
                )
            )
        if self.config.has_gemini_api():
            models.append(
                Model(
                    id="gemini-2.5-flash",
                    object="model",
                    created=int(time.time()),
                    owned_by="google",
                )
            )

        return models

    async def create_completion(
        self,
        messages: List[ChatMessage],
        model: str,
        user_id: str,
        temperature: Optional[float] = None,
        max_tokens: Optional[int] = None,
        stream: Optional[bool] = False,
    ) -> Dict[str, Any]:
        """Create a chat completion"""
        try:
            temp = temperature if temperature is not None else self.config.temperature
            streaming = stream if stream is not None else False
            max_tok = max_tokens if max_tokens is not None else self.config.max_tokens

            conversation_text = self._format_messages_for_model(messages)

            if streaming:
                return await self._create_streaming_completion(
                    conversation_text, model, temp, max_tok, user_id
                )
            else:
                return await self._create_standard_completion(
                    conversation_text, model, temp, max_tok, user_id
                )

        except Exception as e:
            print(f"Error creating completion: {e}")
            raise e

    def _format_messages_for_model(
        self,
        messages: List[ChatMessage],
        context: ConversationContext = ConversationContext.GENERAL,
        level: LearningLevel = LearningLevel.INTERMEDIATE,
        topic: Optional[str] = None,
        learning_goals: Optional[List[str]] = None,
        user_interests: Optional[List[str]] = None,
    ) -> str:
        """Convert OpenAI message format to text with dynamic prompts"""
        formatted_messages = []

        for msg in messages:
            if msg.role.value == "system":
                formatted_messages.append(f"System: {msg.content}")
            elif msg.role.value == "user":
                formatted_messages.append(f"Human: {msg.content}")
            elif msg.role.value == "assistant":
                formatted_messages.append(f"Assistant: {msg.content}")

        # Create dynamic system prompt based on context
        system_prompt = create_system_prompt(
            context=context,
            level=level,
            topic=topic,
            learning_goals=learning_goals,
            user_interests=user_interests,
        )
        formatted_messages.insert(0, f"System: {system_prompt}")

        return "\n".join(formatted_messages) + "\nAssistant:"

    async def _create_standard_completion(
        self,
        conversation_text: str,
        model: str,
        temperature: float,
        max_tokens: int,
        user_id: str,
    ) -> Dict[str, Any]:
        """Create standard completion"""
        response_text = await self._generate_english_learning_response(
            conversation_text, temperature
        )

        prompt_tokens = len(conversation_text.split())
        completion_tokens = len(response_text.split())

        return {
            "content": response_text,
            "prompt_tokens": prompt_tokens,
            "completion_tokens": completion_tokens,
            "total_tokens": prompt_tokens + completion_tokens,
        }

    async def _create_streaming_completion(
        self,
        conversation_text: str,
        model: str,
        temperature: float,
        max_tokens: int,
        user_id: str,
    ) -> Dict[str, Any]:
        """Create streaming completion"""
        stream = self._generate_streaming_response(conversation_text, temperature)
        return {"stream": stream}

    async def _generate_english_learning_response(
        self, conversation_text: str, temperature: float
    ) -> str:
        """Generate intelligent English learning response"""
        lines = conversation_text.strip().split("\n")
        user_message = ""
        for line in reversed(lines):
            if line.startswith("Human:"):
                user_message = line.replace("Human:", "").strip()
                break

        if not user_message:
            return "Hello! I'm your AI English conversation partner. How can I help you practice English today? 😊"

        return self._generate_tutoring_response(user_message, temperature)

    def _generate_tutoring_response(self, user_message: str, temperature: float) -> str:
        """Generate contextual tutoring response with dynamic prompts"""
        user_lower = user_message.lower()

        # Detect conversation context based on user message
        context = ConversationContext.GENERAL
        if any(
            word in user_lower for word in ["grammar", "correct", "mistake", "wrong"]
        ):
            context = ConversationContext.GRAMMAR
        elif any(
            word in user_lower
            for word in ["vocabulary", "words", "meaning", "definition"]
        ):
            context = ConversationContext.VOCABULARY
        elif any(
            word in user_lower
            for word in ["pronounce", "pronunciation", "sound", "accent"]
        ):
            context = ConversationContext.PRONUNCIATION
        elif any(
            word in user_lower
            for word in ["business", "work", "job", "office", "career"]
        ):
            context = ConversationContext.BUSINESS_ENGLISH
        elif any(
            word in user_lower for word in ["conversation", "talk", "speak", "practice"]
        ):
            context = ConversationContext.CONVERSATION_PRACTICE

        # Generate appropriate response based on context
        if context == ConversationContext.GRAMMAR:
            return "I'd be happy to help with grammar! **Grammar practice** is essential for fluency. Try writing a sentence, and I'll help you polish it. Remember: practice makes perfect! What specific grammar topic would you like to work on?"
        elif context == ConversationContext.VOCABULARY:
            return "Building **vocabulary** is exciting! Try learning **5 new words** daily and use them in sentences. **Context** is key - learn words in phrases, not isolation. What topics interest you? I can suggest vocabulary for those areas!"
        elif context == ConversationContext.PRONUNCIATION:
            return "**Pronunciation** is so important! Here's a tip: practice with **tongue twisters** and **minimal pairs** (like 'ship' vs 'sheep'). Record yourself speaking and listen back. What specific sounds would you like to practice?"
        elif context == ConversationContext.BUSINESS_ENGLISH:
            return "Work conversations are great practice! **Professional English** uses specific vocabulary. Try describing your typical workday or dream job. What kind of work do you do or want to do?"
        elif context == ConversationContext.CONVERSATION_PRACTICE:
            return "**Conversation practice** is the best way to improve! Let's have a natural chat. Remember: don't worry about perfect grammar - **communication** comes first! What's something interesting that happened to you recently?"

        # Topic-based responses
        if any(word in user_lower for word in ["food", "eat", "cooking", "restaurant"]):
            return "Food is a delicious topic! 🍕 Practice with **cooking verbs** (chop, boil, fry) and **taste adjectives** (savory, spicy, tender). What's your favorite dish? Can you describe how to make it?"
        elif any(
            word in user_lower for word in ["travel", "trip", "vacation", "country"]
        ):
            return "Travel stories are perfect for practice! ✈️ Use **past tense** to describe trips and **future tense** for plans. **Descriptive language** makes travel stories engaging. Where would you love to visit?"

        # Default encouraging response with dynamic encouragement
        import random

        encouragements = get_encouragement_phrases(LearningLevel.INTERMEDIATE)
        encouragement = random.choice(encouragements) if encouragements else "Great!"

        return f"{encouragement} I'd love to hear more about that. **Expanding** on your thoughts is great practice - try adding more details, examples, or your personal opinions. What else can you tell me about this topic? 🤔"

    async def _generate_streaming_response(
        self, conversation_text: str, temperature: float
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """Generate streaming response"""
        full_response = await self._generate_english_learning_response(
            conversation_text, temperature
        )
        words = full_response.split()

        for i, word in enumerate(words):
            chunk = {
                "id": f"chatcmpl-{str(uuid.uuid4())}",
                "object": "chat.completion.chunk",
                "created": int(time.time()),
                "model": "gpt-4-tutor",
                "choices": [
                    {
                        "index": 0,
                        "delta": {"content": word + " "},
                        "finish_reason": None if i < len(words) - 1 else "stop",
                    }
                ],
            }
            yield chunk
            await asyncio.sleep(0.1)  # Simulate realistic streaming delay

    def simple_chat(self, message: str, model: str) -> str:
        """Simple chat interface for basic conversations"""
        return self._generate_tutoring_response(message, 0.7)

    async def create_language_session(
        self,
        user_id: str,
        language: str = "English",
        level: str = "intermediate",
        context: str = "general",
        topic: Optional[str] = None,
        goals: Optional[List[str]] = None,
        interests: Optional[List[str]] = None,
    ) -> LanguageSession:
        """Create a new language learning session with enhanced options"""
        session_id = str(uuid.uuid4())
        session = LanguageSession(
            id=session_id,
            user_id=user_id,
            language=language,
            level=level,
            context=context,
            topic=topic,
            goals=goals or [],
            interests=interests or [],
            created_at=datetime.utcnow().isoformat(),
            last_activity=datetime.utcnow().isoformat(),
        )
        self.sessions_store[session_id] = session
        return session

    def get_conversation_starter(
        self,
        context: str = "general",
        level: str = "intermediate",
        topic: Optional[str] = None,
    ) -> str:
        """Get an appropriate conversation starter"""
        try:
            conv_context = ConversationContext(context)
        except ValueError:
            conv_context = ConversationContext.GENERAL

        try:
            learning_level = LearningLevel(level)
        except ValueError:
            learning_level = LearningLevel.INTERMEDIATE

        return create_conversation_starter(conv_context, learning_level, topic)


# =====================================================
# AUTHENTICATION - Simple JWT-based authentication
# =====================================================

security = HTTPBearer()


async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
):
    """Simple authentication - in production use proper JWT validation"""
    # For simplicity, just return a mock user
    # In production, validate the JWT token properly
    return {"user_id": "demo_user", "username": "demo"}


# =====================================================
# APPLICATION SETUP
# =====================================================

# Global instances
config = Config()
conversation_service = ConversationService(config)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan manager"""
    print("🚀 Starting AI English Speaking Learning Service (Simplified)...")

    await conversation_service.initialize()

    print(f"✅ Service started successfully!")
    print(f"📝 Server running on: http://{config.host}:{config.port}")
    print(f"📖 API docs: http://{config.host}:{config.port}/docs")

    yield

    await conversation_service.cleanup()
    print("✅ Service stopped gracefully")


# Initialize FastAPI application
app = FastAPI(
    title="AI English Speaking Learning Service",
    description="""
    🗣️ **AI-powered English conversation partner for language learning**
    
    A simplified, consolidated service for English conversation practice with AI.
    
    ## Features
    - 🤖 **Multiple AI models** support
    - 🔄 **Streaming responses** 
    - 🔐 **Simple authentication**
    - 🌍 **CORS enabled**
    - 📱 **Mobile-friendly** responses
    
    ## Available Endpoints
    - `/v1/models` - List available models
    - `/v1/chat/completions` - Chat completions (streaming supported)
    - `/v1/language/session` - Create language learning sessions
    - `/conversation` - Simple conversation endpoint
    - `/health` - Health check
    """,
    version="2.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
    lifespan=lifespan,
)

# Configure CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# =====================================================
# ENDPOINTS - All API endpoints in one place
# =====================================================


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "healthy", "timestamp": datetime.utcnow().isoformat()}


@app.get("/v1/models", response_model=ModelList)
async def list_models(user=Depends(get_current_user)):
    """List available models"""
    models = conversation_service.get_available_models()
    return ModelList(data=models)


@app.post("/v1/chat/completions")
async def create_chat_completion(
    request: ChatCompletionRequest, user=Depends(get_current_user)
):
    """Create a chat completion (OpenAI compatible)"""
    try:
        completion = await conversation_service.create_completion(
            messages=request.messages,
            model=request.model,
            user_id=user.get("user_id"),
            temperature=request.temperature,
            max_tokens=request.max_tokens,
            stream=request.stream,
        )

        if request.stream:
            stream_generator = completion.get("stream")
            if stream_generator:

                async def stream_chat_completion():
                    try:
                        async for chunk in stream_generator:
                            chunk_json = json.dumps(chunk, ensure_ascii=False)
                            yield f"data: {chunk_json}\n\n"
                        yield "data: [DONE]\n\n"
                    except Exception as e:
                        error_chunk = {
                            "error": {"message": str(e), "type": "internal_error"}
                        }
                        yield f"data: {json.dumps(error_chunk)}\n\n"

                return StreamingResponse(
                    stream_chat_completion(),
                    media_type="text/plain",
                    headers={
                        "Cache-Control": "no-cache",
                        "Connection": "keep-alive",
                        "Content-Type": "text/plain; charset=utf-8",
                    },
                )

        # Standard response
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


@app.post("/v1/language/session", response_model=LanguageSession)
async def create_language_session(
    language: str = "English",
    level: str = "intermediate",
    context: str = "general",
    topic: Optional[str] = None,
    goals: Optional[str] = None,  # comma-separated string
    interests: Optional[str] = None,  # comma-separated string
    user=Depends(get_current_user),
):
    """Create a new enhanced language learning session"""
    # Parse comma-separated strings into lists
    goals_list = [g.strip() for g in goals.split(",")] if goals else []
    interests_list = [i.strip() for i in interests.split(",")] if interests else []

    session = await conversation_service.create_language_session(
        user_id=user.get("user_id"),
        language=language,
        level=level,
        context=context,
        topic=topic,
        goals=goals_list,
        interests=interests_list,
    )
    return session


@app.post("/conversation")
async def simple_conversation(
    message: str, model: str = "gpt-4-tutor", user=Depends(get_current_user)
):
    """Simple conversation endpoint"""
    response = conversation_service.simple_chat(message, model)
    return {"response": response, "model": model}


@app.get("/v1/conversation/starter")
async def get_conversation_starter(
    context: str = "general",
    level: str = "intermediate",
    topic: Optional[str] = None,
    user=Depends(get_current_user),
):
    """Get an appropriate conversation starter based on context and level"""
    starter = conversation_service.get_conversation_starter(context, level, topic)
    return {"starter": starter, "context": context, "level": level, "topic": topic}


# =====================================================
# DEVELOPMENT SERVER
# =====================================================

if __name__ == "__main__":
    print("🔧 Starting development server...")
    print(f"📍 Environment: {'DEBUG' if config.debug else 'PRODUCTION'}")
    print(f"🌐 Host: {config.host}")
    print(f"🔌 Port: {config.port}")

    uvicorn.run(
        "app:app",
        host=config.host,
        port=config.port,
        reload=config.debug,
        log_level=config.log_level.lower(),
        access_log=True,
        use_colors=True,
    )
