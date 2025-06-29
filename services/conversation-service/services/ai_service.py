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
        """Generate educational response based on user message"""

        message_lower = user_message.lower()

        # 数学相关
        if any(
            word in message_lower
            for word in [
                "数学",
                "方程",
                "函数",
                "微积分",
                "代数",
                "math",
                "equation",
                "function",
            ]
        ):
            responses = [
                "让我们一步步来理解这个数学概念。首先，你能告诉我你对这个问题的初步想法吗？",
                "数学是一个非常有趣的领域！为了更好地帮助你，你能具体说明你在哪个部分遇到困难了吗？",
                "很好的问题！让我们从基础开始。你觉得解决这类问题的第一步应该是什么？",
            ]

        # 科学相关
        elif any(
            word in message_lower
            for word in [
                "物理",
                "化学",
                "生物",
                "physics",
                "chemistry",
                "biology",
                "光合作用",
                "photosynthesis",
            ]
        ):
            responses = [
                "科学概念最好通过实例来理解。让我给你举个日常生活中的例子来说明这个概念...",
                "这是一个很重要的科学概念！你能先描述一下你已经知道的相关内容吗？",
                "为了更好地理解这个现象，我们可以从它的基本原理开始。你觉得什么是最重要的因素？",
            ]

        # 编程相关
        elif any(
            word in message_lower
            for word in ["编程", "代码", "programming", "code", "python", "javascript"]
        ):
            responses = [
                "编程的关键是理解逻辑和步骤。让我们分解这个问题，一步一步来解决。",
                "很好的编程问题！你能先告诉我你想要实现什么功能吗？",
                "编程时最重要的是思路清晰。我们先从伪代码开始，你觉得第一步应该做什么？",
            ]

        # 语言学习
        elif any(
            word in message_lower
            for word in ["语法", "单词", "grammar", "vocabulary", "英语", "english"]
        ):
            responses = [
                "语言学习需要多练习。让我们从这个具体的例子开始，你能试着造个句子吗？",
                "理解语法规则很重要，但更重要的是在实际中应用。你想先从哪个方面开始？",
                "语言学习的秘诀是多使用。你能用刚学的这个知识点来回答我一个问题吗？",
            ]

        # 通用问候
        elif any(
            word in message_lower for word in ["你好", "hello", "hi", "帮助", "help"]
        ):
            responses = [
                "你好！我是你的AI学习助手。我可以帮你学习各种学科，包括数学、科学、编程等。你今天想学什么？",
                "很高兴见到你！我专门帮助学生学习和理解各种概念。你有什么学习目标吗？",
                "欢迎！我是一位AI教师，擅长用苏格拉底式教学法帮助学生思考。你遇到什么学习挑战了吗？",
            ]

        # 默认响应
        else:
            responses = [
                "这是一个很有趣的问题！让我们一起探索一下。你能分享更多的背景信息吗？",
                "我理解你的疑问。为了给你最好的帮助，你能告诉我你的学习目标是什么吗？",
                "让我们一步步来分析这个问题。首先，你觉得关键点在哪里？",
                "很好的问题！我建议我们从基础概念开始。你对相关的基本知识了解多少？",
            ]

        # 根据 temperature 选择响应的随机性
        import random

        if temperature > 0.7:
            # 高 temperature，更有创意的响应
            selected_response = random.choice(responses)
            # 添加一些变化
            endings = [
                " 我很期待听到你的想法！",
                " 记住，学习是一个过程，我们慢慢来。",
                " 你的好奇心很棒，这是学习的关键！",
                " 让我们一起探索这个有趣的领域。",
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
        """Get the system prompt for AI tutoring"""
        return """You are an expert AI tutor designed to help students learn effectively. Your role is to:

1. **Understand the student's level**: Adapt your explanations to their knowledge level
2. **Ask guiding questions**: Help students discover answers rather than giving them directly
3. **Provide clear explanations**: Use simple language and examples when needed
4. **Encourage critical thinking**: Challenge students to think deeper about concepts
5. **Be patient and supportive**: Create a positive learning environment
6. **Check understanding**: Regularly verify that students understand before moving forward
7. **Provide practical examples**: Use real-world applications to illustrate concepts

Guidelines:
- Always maintain a friendly, encouraging tone
- Break down complex topics into manageable parts
- Use the Socratic method when appropriate
- Celebrate student progress and breakthroughs
- Adapt to different learning styles
- If a student seems frustrated, provide more guidance and support
- End responses with a question or suggestion for the next step when appropriate

Remember: Your goal is to facilitate learning, not just provide answers."""

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
