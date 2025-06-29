# AI English Speaking Learning Service (Simplified)

🗣️ **A consolidated AI-powered English conversation partner for language learning**

This is a **simplified, single-file** FastAPI service that provides intelligent English conversation practice. All functionality has been consolidated into one main file for easier development and maintenance.

## ✨ Features

- 🤖 **Multiple AI model support** (Doubao, DeepSeek, Gemini + local models)
- 🔄 **Streaming responses** with Server-Sent Events
- 🔐 **Simple JWT authentication**
- 🌍 **CORS enabled** for web applications
- 📱 **Mobile-friendly** JSON responses
- 🎯 **OpenAI-compatible API** endpoints
- 📝 **Intelligent tutoring responses** with contextual guidance

## 🏗️ Simplified Architecture

```
conversation-service/
├── app.py                    # 🎯 MAIN FILE - Everything in one place!
├── prompt.py                 # 🎭 PROMPT SYSTEM - Dynamic prompt generation
├── pyproject.toml            # Dependencies and configuration
├── README.md                 # This file
├── Dockerfile                # Container configuration
├── test_simple.py            # Simple API testing
├── example_prompt_usage.py   # Prompt system examples
└── tests/                    # Test files (optional)
```

**That's it!** No complex directory structures, no scattered files. Core functionality in `app.py`, intelligent prompts in `prompt.py`.

## 🎭 Dynamic Prompt System

The service now includes a sophisticated prompt management system (`prompt.py`) that creates context-aware, personalized prompts for different learning scenarios.

### Key Features

- **🎯 Context-Aware**: Different prompts for grammar, vocabulary, pronunciation, business English, etc.
- **📊 Level-Adaptive**: Prompts adjust based on learner level (beginner, intermediate, advanced)
- **🎪 Topic-Focused**: Specialized prompts for specific topics (travel, food, work, etc.)
- **👤 Personalized**: Incorporates user goals and interests
- **🎭 Roleplay Support**: Interactive scenario-based learning
- **📈 Assessment-Ready**: Prompts for skill evaluation

### Example Usage

```python
from prompt import create_system_prompt, ConversationContext, LearningLevel

# Create a grammar-focused prompt for beginners
grammar_prompt = create_system_prompt(
    context=ConversationContext.GRAMMAR,
    level=LearningLevel.BEGINNER,
    topic="past tense",
    learning_goals=["Master simple past tense", "Practice irregular verbs"],
    user_interests=["movies", "travel"]
)

# Generate conversation starters
starter = create_conversation_starter(
    ConversationContext.VOCABULARY,
    LearningLevel.INTERMEDIATE,
    "technology"
)
```

### Available Contexts

- `GENERAL` - General conversation practice
- `GRAMMAR` - Grammar-focused learning
- `VOCABULARY` - Vocabulary building
- `PRONUNCIATION` - Pronunciation improvement
- `CONVERSATION_PRACTICE` - Natural conversation flow
- `BUSINESS_ENGLISH` - Professional communication
- `ACADEMIC_ENGLISH` - Scholarly language skills
- `CASUAL_CHAT` - Informal conversation

### Enhanced API Endpoints

| Endpoint                   | Method | Description                          |
| -------------------------- | ------ | ------------------------------------ |
| `/health`                  | GET    | Health check                         |
| `/v1/models`               | GET    | List available models                |
| `/v1/chat/completions`     | POST   | OpenAI-compatible chat completions   |
| `/v1/language/session`     | POST   | Create enhanced learning session     |
| `/v1/conversation/starter` | GET    | Get contextual conversation starters |
| `/conversation`            | POST   | Simple conversation endpoint         |

### Enhanced Session Creation

```bash
curl -X POST "http://localhost:8000/v1/language/session" \
  -H "Authorization: Bearer demo-token" \
  -d "language=English&level=intermediate&context=business_english&topic=presentations&goals=improve presentation skills,learn business vocabulary&interests=technology,entrepreneurship"
```

### Conversation Starters

```bash
curl "http://localhost:8000/v1/conversation/starter?context=vocabulary&level=intermediate&topic=food" \
  -H "Authorization: Bearer demo-token"
```

## 🚀 Quick Start

### 1. Install Dependencies

```bash
# Using uv (recommended)
uv install

# Or using pip
pip install -e .
```

### 2. Set Environment Variables (Optional)

```bash
# External API keys (all optional - service works without them)
export DOUBAO_API_KEY="your-doubao-key"
export DEEPSEEK_API_KEY="your-deepseek-key"
export GEMINI_API_KEY="your-gemini-key"

# Service configuration
export HOST="0.0.0.0"
export PORT="8000"
export DEBUG="true"
```

### 3. Run the Service

```bash
# Development mode (with auto-reload)
python app.py

# Or using uvicorn directly
uvicorn app:app --reload --host 0.0.0.0 --port 8000

# Or using hatch scripts
hatch run dev
```

The service will start at `http://localhost:8000`

## 📚 API Documentation

Visit `http://localhost:8000/docs` for interactive Swagger documentation.

### Available Endpoints

| Endpoint                   | Method | Description                          |
| -------------------------- | ------ | ------------------------------------ |
| `/health`                  | GET    | Health check                         |
| `/v1/models`               | GET    | List available models                |
| `/v1/chat/completions`     | POST   | OpenAI-compatible chat completions   |
| `/v1/language/session`     | POST   | Create enhanced learning session     |
| `/v1/conversation/starter` | GET    | Get contextual conversation starters |
| `/conversation`            | POST   | Simple conversation endpoint         |

### Example Usage

#### Simple Conversation

```bash
curl -X POST "http://localhost:8000/conversation" \
  -H "Authorization: Bearer demo-token" \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello! I want to practice English."}'
```

#### OpenAI-Compatible Chat

```bash
curl -X POST "http://localhost:8000/v1/chat/completions" \
  -H "Authorization: Bearer demo-token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4-tutor",
    "messages": [
      {"role": "user", "content": "Help me practice English conversation"}
    ],
    "stream": false
  }'
```

#### Streaming Response

```bash
curl -X POST "http://localhost:8000/v1/chat/completions" \
  -H "Authorization: Bearer demo-token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4-tutor",
    "messages": [
      {"role": "user", "content": "Tell me about travel"}
    ],
    "stream": true
  }'
```

## 🔧 Configuration

All configuration is done via environment variables:

| Variable           | Default           | Description                 |
| ------------------ | ----------------- | --------------------------- |
| `HOST`             | `0.0.0.0`         | Server host                 |
| `PORT`             | `8000`            | Server port                 |
| `DEBUG`            | `false`           | Enable debug mode           |
| `TEMPERATURE`      | `0.7`             | AI response temperature     |
| `MAX_TOKENS`       | `2000`            | Maximum response tokens     |
| `DOUBAO_API_KEY`   | -                 | Doubao API key (optional)   |
| `DEEPSEEK_API_KEY` | -                 | DeepSeek API key (optional) |
| `GEMINI_API_KEY`   | -                 | Gemini API key (optional)   |
| `JWT_SECRET`       | `your-secret-key` | JWT secret for auth         |

## 🤖 Available Models

The service supports multiple AI models:

- **`gpt-4-tutor`** - Local English tutoring model (default)
- **`local-tutor`** - Alternative local model
- **`doubao-seed-1-6-250615`** - Doubao API (requires API key)
- **`deepseek-chat`** - DeepSeek API (requires API key)
- **`gemini-2.5-flash`** - Gemini API (requires API key)

## 🧪 Development

### Code Quality Tools

```bash
# Format code
hatch run format

# Lint code
hatch run lint

# Type checking
hatch run typecheck

# Run tests
hatch run test

# Run tests with coverage
hatch run test-cov
```

### Adding New Features

Since everything is in one file (`app.py`), adding new features is straightforward:

1. **Add new models** - Update the `Role` enum and `ChatMessage` class
2. **Add new endpoints** - Add functions with `@app.get()` or `@app.post()` decorators
3. **Add new business logic** - Extend the `ConversationService` class
4. **Add new configuration** - Update the `Config` class

## 🐳 Docker

```bash
# Build image
docker build -t conversation-service .

# Run container
docker run -p 8000:8000 conversation-service

# With environment variables
docker run -p 8000:8000 \
  -e DOUBAO_API_KEY="your-key" \
  -e DEBUG="true" \
  conversation-service
```

## 🔐 Authentication

The service uses simple JWT authentication:

- For development: Any bearer token works (returns demo user)
- For production: Implement proper JWT validation in the `get_current_user()` function

Example authorization header:

```
Authorization: Bearer demo-token
```

## 🎯 Why Simplified?

This simplified architecture offers several benefits:

✅ **Easier to understand** - Everything in one file  
✅ **Faster development** - No complex imports or module paths  
✅ **Easier debugging** - All code in one place  
✅ **Simpler deployment** - Single file to manage  
✅ **Better for learning** - Clear code structure  
✅ **Reduced complexity** - No over-engineering

## 📝 Migration from Complex Structure

If you're migrating from the old complex structure:

- **Old**: `app/api/v1/endpoints/chat.py` → **New**: `app.py` (chat endpoints section)
- **Old**: `services/ai_service.py` → **New**: `app.py` (ConversationService class)
- **Old**: `models/openai_models.py` → **New**: `app.py` (models section)
- **Old**: `config/settings.py` → **New**: `app.py` (Config class)
- **Old**: `main.py` → **New**: `app.py` (application setup section)

## 🤝 Contributing

1. Edit `app.py` directly
2. Run tests: `hatch run test`
3. Format code: `hatch run format`
4. Submit PR

## 📄 License

MIT License - see LICENSE file for details.

---

🎉 **Happy English learning!** This simplified service makes it easy to build AI-powered conversation practice applications.
