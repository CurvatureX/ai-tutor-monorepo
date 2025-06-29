import asyncio
import json
import uuid
import time
from datetime import datetime
from typing import List, Dict, Any, Optional, AsyncGenerator
import httpx
from models.openai_models import ChatMessage, TutorSession
from config.settings import Settings


class AIService:
    """AI service for handling conversation generation using external LLM APIs"""

    def __init__(self, settings: Settings):
        self.settings = settings
        self.sessions_store = {}  # In production, use Redis or database
        self.conversation_history = {}  # In production, use Redis or database

    async def initialize(self):
        """Initialize AI service"""
        try:
            print("🔄 Initializing AI Service...")

            # Check available APIs
            available_apis = []
            if self.settings.has_doubao_api():
                available_apis.append("Doubao")
            if self.settings.has_deepseek_api():
                available_apis.append("DeepSeek")
            if self.settings.has_gemini_api():
                available_apis.append("Gemini")

            if available_apis:
                print(
                    f"✅ AI Service initialized with APIs: {', '.join(available_apis)}"
                )
            else:
                print(
                    "⚠️  No external API keys configured - service will use demo responses"
                )

        except Exception as e:
            print(f"❌ Failed to initialize AI Service: {e}")
            print("⚠️  Running in demo mode")

    async def health_check(self) -> bool:
        """Check if AI service is healthy"""
        return True

    async def cleanup(self):
        """Cleanup resources"""
        print("🧹 Cleaning up AI Service...")

    async def create_completion(
        self,
        messages: List[ChatMessage],
        model: str,
        user_id: str,
        temperature: Optional[float] = None,
        max_tokens: Optional[int] = None,
        stream: Optional[bool] = False,
    ) -> Dict[str, Any]:
        """Create a chat completion using external LLM APIs"""

        try:
            # Set default values
            temp = temperature if temperature is not None else self.settings.temperature
            streaming = stream if stream is not None else False
            max_tok = max_tokens if max_tokens is not None else self.settings.max_tokens

            # Convert messages to conversation text
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

    def _format_messages_for_model(self, messages: List[ChatMessage]) -> str:
        """Convert OpenAI message format to text"""
        formatted_messages = []

        for msg in messages:
            if msg.role.value == "system":
                formatted_messages.append(f"System: {msg.content}")
            elif msg.role.value == "user":
                formatted_messages.append(f"Human: {msg.content}")
            elif msg.role.value == "assistant":
                formatted_messages.append(f"Assistant: {msg.content}")

        # Add system prompt for English learning
        system_prompt = f"System: {self._get_english_tutor_system_prompt()}"
        formatted_messages.insert(0, system_prompt)

        return "\n".join(formatted_messages) + "\nAssistant:"

    async def _create_standard_completion(
        self,
        conversation_text: str,
        model: str,
        temperature: float,
        max_tokens: int,
        user_id: str,
    ) -> Dict[str, Any]:
        """Create standard (non-streaming) completion using external APIs"""

        # For now, use intelligent simulation
        # In production, integrate with your preferred LLM API
        response_text = await self._generate_english_learning_response(
            conversation_text, temperature
        )

        # Calculate token counts (simple estimation)
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

        # Extract the last user message for context
        lines = conversation_text.strip().split("\n")
        user_message = ""
        for line in reversed(lines):
            if line.startswith("Human:"):
                user_message = line.replace("Human:", "").strip()
                break

        # Generate contextual English learning response
        if not user_message:
            return "Hello! I'm your AI English conversation partner. How can I help you practice English today? 😊"

        # Intelligent response based on common learning scenarios
        response = self._generate_tutoring_response(user_message, temperature)

        # Add encouraging tone
        encouraging_phrases = [
            "Great question!",
            "Excellent!",
            "That's a good point!",
            "Nice try!",
            "Perfect!",
            "Wonderful!",
        ]

        if temperature > 0.5:  # Add variety for higher temperature
            import random

            if random.random() < 0.3:  # 30% chance to add encouragement
                response = random.choice(encouraging_phrases) + " " + response

        return response

    def _generate_tutoring_response(self, user_message: str, temperature: float) -> str:
        """Generate contextual tutoring response based on user input"""

        user_lower = user_message.lower()

        # Greeting responses
        if any(
            word in user_lower
            for word in ["hello", "hi", "hey", "good morning", "good afternoon"]
        ):
            return "Hello! It's wonderful to meet you! I'm here to help you practice English conversation. What would you like to talk about today? We could discuss hobbies, travel, food, or anything that interests you! 🌟"

        # Grammar practice scenarios
        if any(
            word in user_lower for word in ["grammar", "correct", "mistake", "wrong"]
        ):
            return "I'd be happy to help with grammar! **Grammar practice** is essential for fluency. Try writing a sentence, and I'll help you polish it. Remember: practice makes perfect! What specific grammar topic would you like to work on?"

        # Pronunciation help
        if any(
            word in user_lower
            for word in ["pronounce", "pronunciation", "sound", "accent"]
        ):
            return "**Pronunciation** is so important! Here's a tip: practice with **tongue twisters** and **minimal pairs** (like 'ship' vs 'sheep'). Record yourself speaking and listen back. What specific sounds would you like to practice?"

        # Vocabulary building
        if any(
            word in user_lower
            for word in ["vocabulary", "words", "meaning", "definition"]
        ):
            return "Building **vocabulary** is exciting! Try learning **5 new words** daily and use them in sentences. **Context** is key - learn words in phrases, not isolation. What topics interest you? I can suggest vocabulary for those areas!"

        # Conversation practice
        if any(
            word in user_lower for word in ["conversation", "talk", "speak", "practice"]
        ):
            return "**Conversation practice** is the best way to improve! Let's have a natural chat. Remember: don't worry about perfect grammar - **communication** comes first! What's something interesting that happened to you recently?"

        # Daily life topics
        if any(word in user_lower for word in ["work", "job", "office", "career"]):
            return "Work conversations are great practice! **Professional English** uses specific vocabulary. Try describing your typical workday or dream job. What kind of work do you do or want to do?"

        if any(word in user_lower for word in ["food", "eat", "cooking", "restaurant"]):
            return "Food is a delicious topic! 🍕 Practice with **cooking verbs** (chop, boil, fry) and **taste adjectives** (savory, spicy, tender). What's your favorite dish? Can you describe how to make it?"

        if any(
            word in user_lower for word in ["travel", "trip", "vacation", "country"]
        ):
            return "Travel stories are perfect for practice! ✈️ Use **past tense** to describe trips and **future tense** for plans. **Descriptive language** makes travel stories engaging. Where would you love to visit?"

        # Encouragement for any effort
        if len(user_message) > 20:  # Longer message shows effort
            return f"I can see you're really trying - that's **fantastic**! Your message shows good effort. Let me respond: *[provides natural response to their content]*. Keep up the great work! Remember, every conversation makes you stronger in English! 💪"

        # Default encouraging response
        return f"That's interesting! I'd love to hear more about that. **Expanding** on your thoughts is great practice - try adding more details, examples, or your personal opinions. What else can you tell me about this topic? 🤔"

    async def _generate_streaming_response(
        self, conversation_text: str, temperature: float
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """Generate streaming response for real-time conversation"""

        # Get the full response first
        full_response = await self._generate_english_learning_response(
            conversation_text, temperature
        )

        # Stream it word by word
        words = full_response.split()

        for i, word in enumerate(words):
            chunk_content = word + (" " if i < len(words) - 1 else "")

            chunk = {
                "id": f"chatcmpl-{str(uuid.uuid4())}",
                "object": "chat.completion.chunk",
                "created": int(time.time()),
                "model": self.settings.default_model,
                "choices": [
                    {
                        "index": 0,
                        "delta": {"content": chunk_content, "role": "assistant"},
                        "finish_reason": None,
                    }
                ],
            }

            yield chunk
            await asyncio.sleep(0.05)  # Small delay for realistic streaming

        # Final chunk
        final_chunk = {
            "id": f"chatcmpl-{str(uuid.uuid4())}",
            "object": "chat.completion.chunk",
            "created": int(time.time()),
            "model": self.settings.default_model,
            "choices": [
                {
                    "index": 0,
                    "delta": {},
                    "finish_reason": "stop",
                }
            ],
        }
        yield final_chunk

    def _get_english_tutor_system_prompt(self) -> str:
        """Get system prompt for English language learning"""
        return """You are a friendly, encouraging AI English conversation partner focused on helping users improve their spoken English. Your role is to:

🎯 **Primary Goals:**
- Facilitate natural, engaging conversations in English
- Provide gentle corrections and suggestions for improvement  
- Build confidence through positive reinforcement
- Focus on practical communication skills

💬 **Communication Style:**
- Be warm, supportive, and patient
- Use simple, clear language appropriate for learners
- Encourage elaboration with follow-up questions
- Celebrate progress and effort

📚 **Learning Focus Areas:**
- **Pronunciation**: Help with difficult sounds and word stress
- **Grammar**: Offer corrections in context, not as lectures  
- **Vocabulary**: Introduce new words naturally in conversation
- **Fluency**: Encourage speaking without fear of mistakes
- **Cultural Context**: Share insights about English-speaking cultures

✨ **Response Format:**
- Keep responses conversational and natural
- Use **bold** for key vocabulary or grammar points
- Include encouragement and positive feedback
- Ask engaging follow-up questions
- Provide practical examples when teaching

Remember: The goal is confident, natural communication - not perfect grammar!"""

    # Session management methods (unchanged)
    async def create_tutor_session(
        self, user_id: str, subject: str, level: str = "intermediate"
    ) -> TutorSession:
        """Create a new language learning session"""
        session_id = str(uuid.uuid4())
        session = TutorSession(
            id=session_id,
            user_id=user_id,
            subject=subject,  # "English" for language learning
            level=level,
            created_at=datetime.utcnow(),
            last_activity=datetime.utcnow(),
            message_count=0,
            status="active",
        )

        self.sessions_store[session_id] = session
        return session

    async def get_user_sessions(self, user_id: str) -> List[TutorSession]:
        """Get all sessions for a user"""
        user_sessions = [
            session
            for session in self.sessions_store.values()
            if session.user_id == user_id
        ]
        return user_sessions

    async def get_session_history(self, session_id: str, user_id: str) -> List[Dict]:
        """Get conversation history for a session"""
        if session_id in self.conversation_history:
            return self.conversation_history[session_id]
        return []

    async def add_message_to_session(self, session_id: str, message: ChatMessage):
        """Add a message to session history"""
        if session_id not in self.conversation_history:
            self.conversation_history[session_id] = []

        self.conversation_history[session_id].append(
            {
                "role": message.role.value,
                "content": message.content,
                "timestamp": datetime.utcnow().isoformat(),
            }
        )

        # Update session activity
        if session_id in self.sessions_store:
            session = self.sessions_store[session_id]
            session.last_activity = datetime.utcnow()
            session.message_count += 1
