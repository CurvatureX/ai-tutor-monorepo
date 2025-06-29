from pydantic import BaseModel, Field
from typing import List, Optional, Union, Dict, Any
from enum import Enum


class Role(str, Enum):
    SYSTEM = "system"
    USER = "user"
    ASSISTANT = "assistant"
    FUNCTION = "function"


class ChatMessage(BaseModel):
    role: Role
    content: str
    name: Optional[str] = None
    function_call: Optional[Dict[str, Any]] = None


class FunctionCall(BaseModel):
    name: str
    arguments: str


class Function(BaseModel):
    name: str
    description: Optional[str] = None
    parameters: Optional[Dict[str, Any]] = None


class ChatCompletionRequest(BaseModel):
    model: str = Field(default="gpt-4-tutor", description="AI model to use")
    messages: List[ChatMessage] = Field(
        ..., description="List of messages in the conversation"
    )
    temperature: Optional[float] = Field(
        default=0.7, ge=0.0, le=2.0, description="Sampling temperature"
    )
    max_tokens: Optional[int] = Field(
        default=None, ge=1, description="Maximum tokens to generate"
    )
    top_p: Optional[float] = Field(
        default=1.0, ge=0.0, le=1.0, description="Nucleus sampling parameter"
    )
    n: Optional[int] = Field(
        default=1, ge=1, le=128, description="Number of completions to generate"
    )
    stream: Optional[bool] = Field(
        default=False, description="Whether to stream the response"
    )
    stop: Optional[Union[str, List[str]]] = Field(
        default=None, description="Stop sequences"
    )
    presence_penalty: Optional[float] = Field(
        default=0.0, ge=-2.0, le=2.0, description="Presence penalty"
    )
    frequency_penalty: Optional[float] = Field(
        default=0.0, ge=-2.0, le=2.0, description="Frequency penalty"
    )
    logit_bias: Optional[Dict[str, float]] = Field(
        default=None, description="Logit bias"
    )
    user: Optional[str] = Field(default=None, description="User identifier")
    functions: Optional[List[Function]] = Field(
        default=None, description="Available functions"
    )
    function_call: Optional[Union[str, Dict[str, str]]] = Field(
        default=None, description="Function call preference"
    )


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


class ChatCompletionStreamChoice(BaseModel):
    index: int
    delta: Dict[str, Any]
    finish_reason: Optional[str] = None


class ChatCompletionStreamResponse(BaseModel):
    id: str
    object: str = "chat.completion.chunk"
    created: int
    model: str
    choices: List[ChatCompletionStreamChoice]


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
    language: str = "English"  # 默认英语，后续支持其他语言
    level: str = "intermediate"  # 口语水平：beginner, intermediate, advanced
    goals: List[str] = []  # 学习目标，如 ["IELTS speaking", "Business communication"]
    created_at: str
    last_activity: str
    message_count: int = 0
    status: str = "active"


# 保持向后兼容
TutorSession = LanguageSession


class SessionHistory(BaseModel):
    session_id: str
    messages: List[ChatMessage]
    created_at: str
    updated_at: str
