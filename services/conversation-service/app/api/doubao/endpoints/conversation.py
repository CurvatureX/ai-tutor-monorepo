"""
Doubao API conversation endpoints
"""

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import Optional
from datetime import datetime

from app.api.deps import CurrentUser, DoubaoClientDep
from app.clients.doubao import DoubaoClient

router = APIRouter()


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


@router.post("/conversation")
async def create_conversation(
    request: ConversationRequest,
    user=CurrentUser,
    doubao_client: DoubaoClient = DoubaoClientDep
):
    """
    创建对话 - Doubao API endpoint

    Args:
        request: 包含消息、用户ID和模型的请求体
        user: 通过JWT认证的用户信息

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


@router.post("/conversation/multimodal")
async def create_multimodal_conversation(
    request: MultimodalConversationRequest,
    user=CurrentUser,
    doubao_client: DoubaoClient = DoubaoClientDep
):
    """
    创建多模态对话 - Doubao API endpoint

    Args:
        request: 包含文本、图片URL、用户ID和模型的请求体
        user: 通过JWT认证的用户信息

    Returns:
        包含AI回复的响应
    """
    try:
        print(f"🖼️ Processing multimodal conversation from user {request.user_id}")

        # 调用Doubao多模态客户端
        response_text = doubao_client.multimodal_chat(
            request.text, request.image_url, request.model or "doubao-seed-1-6-250615"
        )
        response = {"content": response_text}

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


@router.post("/test/doubao")
async def test_doubao_api(
    user=CurrentUser,
    doubao_client: DoubaoClient = DoubaoClientDep
):
    """
    测试Doubao API连接

    Args:
        user: 通过JWT认证的用户信息

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
