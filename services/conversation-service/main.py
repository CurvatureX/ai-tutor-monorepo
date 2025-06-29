"""
AI English Speaking Learning Service - Main Application

AI-powered English conversation partner for language learning.
Supports both OpenAI-compatible API and Doubao native API.

Architecture: Modular design with clear separation of concerns
- API Layer: Organized endpoints for different features
- Core Layer: Security, authentication, and shared utilities
- Business Layer: AI services and conversation logic
- Client Layer: External service integrations

Version: 1.0.0 (Refactored)
"""

import uvicorn
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

# Import configurations and core modules
from config.settings import Settings
from app.core.security import init_security
from app.api.deps import init_dependencies

# Import services
from services.ai_service import AIService
from services.auth_service import AuthService
from app.clients.doubao import DoubaoClient

# Import API routers
from app.api.v1.router import router as v1_router
from app.api.doubao.router import router as doubao_router


@asynccontextmanager
async def lifespan(app: FastAPI):
    """
    Application lifespan manager

    Replaces deprecated startup/shutdown events with modern lifespan pattern.
    Handles service initialization and cleanup gracefully.
    """
    # === STARTUP ===
    print("🚀 Starting AI English Speaking Learning Service...")

    # Load configuration
    settings = Settings()

    # Initialize core services
    ai_service = AIService(settings)
    await ai_service.initialize()

    auth_service = AuthService(settings)
    doubao_client = DoubaoClient()

    # Initialize security and dependencies
    init_security(settings)
    init_dependencies(settings, ai_service, doubao_client)

    # Startup complete
    print(f"✅ Service started successfully!")
    print(f"📝 Server running on: http://{settings.host}:{settings.port}")
    print(f"🎯 APIs: OpenAI-compatible + Doubao native")
    print(f"🗣️ Ready for English conversation practice!")
    print(f"📖 API docs: http://{settings.host}:{settings.port}/docs")

    # Show configured APIs
    configured_apis = []
    if settings.has_doubao_api():
        configured_apis.append("Doubao ✅")
    if settings.has_deepseek_api():
        configured_apis.append("DeepSeek ✅")
    if settings.has_gemini_api():
        configured_apis.append("Gemini ✅")

    if configured_apis:
        print(f"🔌 External APIs: {', '.join(configured_apis)}")
    else:
        print("⚠️  No external API keys configured - using demo responses")

    yield

    # === SHUTDOWN ===
    print("👋 Shutting down AI English Speaking Learning Service...")
    await ai_service.cleanup()
    print("✅ Service stopped gracefully")


# Initialize FastAPI application
app = FastAPI(
    title="AI English Speaking Learning Service",
    description="""
    🗣️ **AI-powered English conversation partner for language learning**
    
    This service provides intelligent conversation practice with focus on:
    - **Pronunciation improvement** 🎵
    - **Grammar practice** 📝  
    - **Fluency building** 💫
    - **Natural conversation** 💬
    
    ## Supported APIs
    
    ### OpenAI-Compatible API
    - `/v1/models` - List available models
    - `/v1/chat/completions` - Chat completions (streaming supported)
    - `/v1/language/session` - Language learning sessions
    
    ### Doubao Native API  
    - `/conversation` - Simple conversations
    - `/conversation/multimodal` - Image + text conversations
    - `/test/doubao` - API connection testing
    
    ## Authentication
    All endpoints require JWT authentication via Bearer token.
    
    ## Features
    - 🤖 **Multiple AI models** (Local + Doubao)
    - 🔄 **Streaming responses** (Server-Sent Events)
    - 🔐 **Secure authentication** (JWT)
    - 🌍 **CORS enabled** for web applications
    - 📱 **Mobile-friendly** JSON responses
    """,
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
    lifespan=lifespan,
    contact={
        "name": "AI Tutor Team",
        "email": "support@ai-tutor.com",
    },
    license_info={
        "name": "MIT License",
        "url": "https://opensource.org/licenses/MIT",
    },
)

# Configure CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure appropriately for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include API routers
app.include_router(v1_router, tags=["v1-api"])
app.include_router(doubao_router, tags=["doubao-api"])


# Development server entry point
if __name__ == "__main__":
    # Load settings for development
    settings = Settings()

    print("🔧 Starting development server...")
    print(f"📍 Environment: {'DEBUG' if settings.debug else 'PRODUCTION'}")
    print(f"🌐 Host: {settings.host}")
    print(f"🔌 Port: {settings.port}")

    # Run the server
    uvicorn.run(
        "main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
        log_level=settings.log_level.lower(),
        access_log=True,
        use_colors=True,
    )
