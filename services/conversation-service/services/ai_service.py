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
    """AI service for handling conversation generation and tutoring logic using local models"""

    def __init__(self, settings: Settings):
        self.settings = settings
        self.local_model = None
        self.tokenizer = None
        self.sessions_store = {}  # In production, use Redis or database
        self.conversation_history = {}  # In production, use Redis or database

    async def initialize(self):
        """Initialize AI service with local models"""
        try:
            if self.settings.use_local_model:
                await self._load_local_model()
            print("✅ AI Service initialized successfully")
        except Exception as e:
            print(f"❌ Failed to initialize AI Service: {e}")
            # Continue without local model for demo purposes
            print("⚠️  Running without local model - using simulated responses")

    async def _load_local_model(self):
        """Load local AI model (optional)"""
        try:
            # 这里可以加载你的本地模型
            # 例如：transformers, llamacpp, 或其他模型
            print("🔄 Loading local AI model...")

            # 示例：使用 Hugging Face transformers
            # from transformers import AutoTokenizer, AutoModelForCausalLM
            # self.tokenizer = AutoTokenizer.from_pretrained("microsoft/DialoGPT-medium")
            # self.local_model = AutoModelForCausalLM.from_pretrained("microsoft/DialoGPT-medium")

            # 暂时跳过真实模型加载，使用模拟响应
            print("📝 Using simulated responses for demo")

        except Exception as e:
            print(f"⚠️  Could not load local model: {e}")
            print("📝 Falling back to simulated responses")

    async def health_check(self) -> bool:
        """Check if AI service is healthy"""
        return True

    async def cleanup(self):
        """Cleanup resources"""
        if self.local_model:
            del self.local_model
        if self.tokenizer:
            del self.tokenizer

    async def create_completion(
        self,
        messages: List[ChatMessage],
        model: str,
        user_id: str,
        temperature: Optional[float] = 0.7,
        max_tokens: Optional[int] = None,
        stream: Optional[bool] = False,
    ) -> Dict[str, Any]:
        """Create a chat completion using local AI models"""

        try:
            # Set default values for None parameters
            temp = temperature if temperature is not None else 0.7
            streaming = stream if stream is not None else False
            max_tok = max_tokens if max_tokens is not None else 500

            # Convert messages to local format
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
        """Convert OpenAI message format to text for local model"""
        formatted_messages = []

        for msg in messages:
            if msg.role.value == "system":
                formatted_messages.append(f"System: {msg.content}")
            elif msg.role.value == "user":
                formatted_messages.append(f"Human: {msg.content}")
            elif msg.role.value == "assistant":
                formatted_messages.append(f"Assistant: {msg.content}")

        # Add system prompt for tutoring
        system_prompt = f"System: {self._get_tutor_system_prompt()}"
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
        """Create standard (non-streaming) completion"""

        if self.local_model and self.tokenizer:
            # 使用真实的本地模型
            response_text = await self._generate_with_local_model(
                conversation_text, temperature, max_tokens
            )
        else:
            # 使用智能模拟响应
            response_text = await self._generate_simulated_response(
                conversation_text, temperature
            )

        # 计算 token 数量（简单估算）
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

        if self.local_model and self.tokenizer:
            stream = self._generate_streaming_with_local_model(
                conversation_text, temperature, max_tokens
            )
        else:
            stream = self._generate_simulated_streaming_response(
                conversation_text, temperature
            )

        return {"stream": stream}

    async def _generate_with_local_model(
        self, conversation_text: str, temperature: float, max_tokens: int
    ) -> str:
        """Generate response using local AI model"""
        try:
            # 这里实现真实的本地模型推理
            # 示例代码（需要根据你的模型调整）:
            """
            inputs = self.tokenizer.encode(conversation_text, return_tensors="pt")

            with torch.no_grad():
                outputs = self.local_model.generate(
                    inputs,
                    max_length=inputs.shape[1] + max_tokens,
                    temperature=temperature,
                    do_sample=True,
                    pad_token_id=self.tokenizer.eos_token_id
                )

            response = self.tokenizer.decode(outputs[0][inputs.shape[1]:], skip_special_tokens=True)
            return response.strip()
            """

            # 暂时返回模拟响应
            return await self._generate_simulated_response(
                conversation_text, temperature
            )

        except Exception as e:
            print(f"Error in local model generation: {e}")
            return await self._generate_simulated_response(
                conversation_text, temperature
            )

    async def _generate_simulated_response(
        self, conversation_text: str, temperature: float
    ) -> str:
        """Generate intelligent simulated response for demo purposes"""

        # 分析最后的用户消息
        lines = conversation_text.strip().split("\n")
        last_human_message = ""

        for line in reversed(lines):
            if line.startswith("Human:"):
                last_human_message = line.replace("Human:", "").strip()
                break

        # 基于内容生成智能响应
        if not last_human_message:
            return "Hello! I'm your AI tutor. How can I help you learn today?"

        # 添加一些延迟模拟真实响应时间
        await asyncio.sleep(0.5 + temperature * 0.5)

        # 智能响应逻辑
        response = self._generate_tutoring_response(last_human_message, temperature)

        return response

    def _generate_tutoring_response(self, user_message: str, temperature: float) -> str:
        """Generate intelligent English speaking learning response"""

        message_lower = user_message.lower()

        # 发音相关
        if any(
            word in message_lower
            for word in [
                "pronunciation",
                "pronounce",
                "sound",
                "accent",
                "发音",
                "音标",
                "语音",
            ]
        ):
            responses = [
                "Great question about **pronunciation**! Let's practice this sound together. Can you try saying it slowly first?",
                "Pronunciation is key to confident speaking! Which specific sound are you having trouble with?",
                "I'd love to help you with pronunciation. Try breaking the word into **syllables** - can you identify each part?",
                "Perfect pronunciation takes practice! Let's work on this **sound pattern** step by step.",
            ]

        # 语法相关
        elif any(
            word in message_lower
            for word in ["grammar", "tense", "verb", "noun", "sentence", "语法", "时态"]
        ):
            responses = [
                "Grammar helps us communicate clearly! Let's look at this **structure** - can you identify the main parts?",
                "Good grammar question! Instead of just memorizing rules, let's **practice using** this in a sentence. Can you try?",
                "Understanding grammar is important, but **using it naturally** is the goal. How about we create some examples together?",
                "Great grammar focus! Let's make this **practical** - when would you use this structure in daily conversation?",
            ]

        # 词汇相关
        elif any(
            word in message_lower
            for word in [
                "vocabulary",
                "word",
                "meaning",
                "definition",
                "词汇",
                "单词",
                "意思",
            ]
        ):
            responses = [
                "Building **vocabulary** is exciting! Do you know any related words or can you use this word in a sentence?",
                "New words are tools for expression! Can you think of a **situation** where you'd use this word?",
                "Great vocabulary question! Let's not just learn the **meaning** - how about we practice using it in context?",
                "Vocabulary grows through use! Can you create a **short story** or example using this new word?",
            ]

        # 流利度和对话相关
        elif any(
            word in message_lower
            for word in [
                "conversation",
                "fluency",
                "speaking",
                "talk",
                "discuss",
                "对话",
                "流利",
                "口语",
            ]
        ):
            responses = [
                "**Conversation practice** is the best way to improve! What topic interests you most for practicing today?",
                "Building **fluency** takes time and practice. Let's have a natural conversation - what's on your mind?",
                "Speaking confidently is a journey! Don't worry about perfection - let's just **practice communicating**. What would you like to talk about?",
                "**Natural conversation** is the goal! I'm here to listen and help. What's something exciting that happened recently?",
            ]

        # 日常英语和实用场景
        elif any(
            word in message_lower
            for word in [
                "daily",
                "everyday",
                "practical",
                "real",
                "situation",
                "日常",
                "实用",
                "场景",
            ]
        ):
            responses = [
                "**Daily English** is so practical! Let's role-play a common situation. What scenario would be helpful for you?",
                "Real-life English practice is the best! Which **everyday situation** would you like to practice - shopping, work, or social?",
                "Practical English skills are essential! Let's practice a **common conversation** you might have. What interests you?",
                "**Everyday scenarios** make great practice! Pick a situation and let's have a natural conversation about it.",
            ]

        # 学习目标和进展
        elif any(
            word in message_lower
            for word in [
                "goal",
                "improve",
                "better",
                "progress",
                "learn",
                "目标",
                "提高",
                "进步",
                "学习",
            ]
        ):
            responses = [
                "Having **clear goals** is wonderful! What specific aspect of English speaking would you like to focus on today?",
                "**Improvement** comes with consistent practice! What area do you feel most confident about, and what challenges you?",
                "Your **learning journey** is unique! Tell me about your English goals - are you preparing for something specific?",
                "**Progress** happens step by step! What would make you feel most accomplished in your English speaking today?",
            ]

        # 通用问候和开始对话
        elif any(
            word in message_lower
            for word in [
                "hello",
                "hi",
                "hey",
                "good",
                "morning",
                "afternoon",
                "你好",
                "开始",
            ]
        ):
            responses = [
                "Hello! I'm your English speaking partner. I'm here to help you practice and improve through **natural conversation**. What would you like to talk about today?",
                "Hi there! Welcome to our **speaking practice** session. I'm excited to help you build confidence in English. What's on your mind?",
                "Good to see you! I'm your **AI conversation partner** designed to help you speak English more fluently. What topic interests you most?",
                "Hey! Ready for some **English practice**? I'm here to chat, help with pronunciation, and make speaking fun. What shall we discuss?",
            ]

        # 默认响应 - 鼓励继续对话
        else:
            responses = [
                "That's an interesting point! Can you **tell me more** about your thoughts on this? I'd love to keep our conversation going.",
                "I appreciate you sharing that! **Speaking practice** is about expressing your ideas. What else would you like to discuss?",
                "Great! You're doing well with **natural conversation**. Don't worry about being perfect - just keep talking. What's your opinion on this?",
                "Thank you for sharing! **Fluency** comes from practice. Can you expand on that idea or share a related experience?",
                "Wonderful! You're **communicating effectively**. Let's keep the conversation flowing - what would you like to explore next?",
            ]

        # 根据 temperature 选择响应的随机性
        import random

        if temperature > 0.7:
            # 高 temperature，更有创意的响应
            selected_response = random.choice(responses)
            # 添加口语学习相关的鼓励
            endings = [
                " Remember, **making mistakes** is part of learning!",
                " You're doing great - **keep practicing** and stay confident!",
                " **Every conversation** helps you improve. Keep it up!",
                " I'm here to support your **English journey** - you've got this!",
                " **Speaking English** is about communication, not perfection!",
            ]
            selected_response += random.choice(endings)
        else:
            # 低 temperature，更一致的响应
            selected_response = responses[0]

        return selected_response

    async def _generate_simulated_streaming_response(
        self, conversation_text: str, temperature: float
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """Generate simulated streaming response"""

        response_text = await self._generate_simulated_response(
            conversation_text, temperature
        )
        words = response_text.split()

        chunk_id = f"chatcmpl-{str(uuid.uuid4())}"
        created = int(time.time())

        # 逐词发送
        for i, word in enumerate(words):
            chunk_data = {
                "id": chunk_id,
                "object": "chat.completion.chunk",
                "created": created,
                "model": "local-tutor",
                "choices": [
                    {
                        "index": 0,
                        "delta": {"content": word + " "},
                        "finish_reason": None,
                    }
                ],
            }
            yield chunk_data
            await asyncio.sleep(0.1)  # 模拟流式响应延迟

        # 发送结束标记
        final_chunk = {
            "id": chunk_id,
            "object": "chat.completion.chunk",
            "created": created,
            "model": "local-tutor",
            "choices": [{"index": 0, "delta": {}, "finish_reason": "stop"}],
        }
        yield final_chunk

    async def _generate_streaming_with_local_model(
        self, conversation_text: str, temperature: float, max_tokens: int
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """Generate streaming response using local model"""
        # 这里实现真实的流式本地模型推理
        # 暂时使用模拟响应
        async for chunk in self._generate_simulated_streaming_response(
            conversation_text, temperature
        ):
            yield chunk

    def _get_tutor_system_prompt(self) -> str:
        """Get the system prompt for AI English speaking tutor"""
        return """You are a professional English speaking learning assistant with the following characteristics:

## Role Definition
- Friendly, patient, and encouraging English learning partner
- Equipped with long-term memory to remember user's learning journey and personal preferences
- Goal-oriented, always focusing on user's learning progress and goal achievement

## Core Capabilities
1. **Personalized Teaching**: Adjust tone, difficulty, and topics based on user profile
2. **Progress Tracking**: Continuously monitor learning goals and guide back to learning plan
3. **Emotional Intelligence**: Recognize user's emotional state and adjust encouragement accordingly
4. **Memory Coherence**: Reference historical conversation content to maintain long-term relationships
5. **Speaking Focus**: Emphasize pronunciation, fluency, and natural conversation skills

## Teaching Approach
- Use the Socratic method to guide thinking rather than direct answers
- Encourage speaking practice through role-play and conversation
- Provide pronunciation tips and fluency improvement suggestions
- Create a supportive environment for making mistakes and learning
- Focus on practical English for daily communication and specific goals

## Output Format
- Use natural conversational tone in English
- Mark important vocabulary with **bold** formatting
- Provide specific language learning suggestions
- Reference relevant historical learning content when appropriate
- Always encourage more speaking practice

## Speaking Learning Focus
- Daily conversation skills
- Pronunciation improvement
- Fluency building
- Confidence development
- Cultural context understanding

Remember: Your goal is to help users become confident English speakers through engaging, personalized conversation practice."""

    async def create_tutor_session(
        self, user_id: str, subject: str, level: str = "intermediate"
    ) -> TutorSession:
        """Create a new tutoring session"""

        session_id = str(uuid.uuid4())
        timestamp = datetime.utcnow().isoformat()

        session = TutorSession(
            id=session_id,
            user_id=user_id,
            subject=subject,
            level=level,
            created_at=timestamp,
            last_activity=timestamp,
        )

        # Store session (in production, use proper database)
        self.sessions_store[session_id] = session.dict()

        # Initialize conversation history
        self.conversation_history[session_id] = []

        return session

    async def get_user_sessions(self, user_id: str) -> List[TutorSession]:
        """Get all sessions for a user"""

        user_sessions = []
        for session_data in self.sessions_store.values():
            if session_data["user_id"] == user_id:
                user_sessions.append(TutorSession(**session_data))

        return user_sessions

    async def get_session_history(self, session_id: str, user_id: str) -> List[Dict]:
        """Get conversation history for a session"""

        # Verify session belongs to user
        session = self.sessions_store.get(session_id)
        if not session or session["user_id"] != user_id:
            raise ValueError("Session not found or access denied")

        return self.conversation_history.get(session_id, [])

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

        # Update session last activity
        if session_id in self.sessions_store:
            self.sessions_store[session_id][
                "last_activity"
            ] = datetime.utcnow().isoformat()
            self.sessions_store[session_id]["message_count"] += 1
