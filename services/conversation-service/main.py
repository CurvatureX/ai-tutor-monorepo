import os
from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import logging
from typing import Optional

from doubao_client import DoubaoClient

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# 加载环境变量
load_dotenv()

# 创建FastAPI应用
app = FastAPI(title="Conversation Service", version="1.0.0")

class Config:
    """配置类，用于管理环境变量"""
    
    def __init__(self):
        self.doubao_api_key = self._get_doubao_api_key()
    
    def _get_doubao_api_key(self) -> str:
        """从环境变量中获取Doubao API Key"""
        api_key = os.getenv("DOUBAO_API_KEY")
        if not api_key:
            raise ValueError("DOUBAO_API_KEY not found in environment variables")
        return api_key
    
    def get_api_key(self) -> str:
        """获取API Key的公共方法"""
        return self.doubao_api_key

# 全局配置实例
config = Config()

# 全局Doubao客户端实例
doubao_client = DoubaoClient(config.get_api_key())

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
    """启动时的初始化事件"""
    logger.info("Conversation Service is starting...")
    logger.info(f"Doubao API Key loaded: {'*' * (len(config.get_api_key()) - 4) + config.get_api_key()[-4:]}")

@app.get("/")
async def root():
    """根路径健康检查"""
    return {"message": "Conversation Service is running"}

@app.get("/health")
async def health_check():
    """健康检查端点"""
    return {
        "status": "healthy",
        "service": "conversation-service",
        "api_key_configured": bool(config.doubao_api_key)
    }

@app.post("/conversation")
async def create_conversation(request: ConversationRequest):
    """创建文本对话的端点"""
    try:
        logger.info(f"Processing text conversation for user: {request.user_id}")
        logger.info(f"Message: {request.message}")
        
        # 使用Doubao客户端进行对话
        response = doubao_client.simple_chat(
            message=request.message,
            model=request.model or "doubao-seed-1-6-250615"
        )
        
        return {
            "message": "Text conversation processed successfully",
            "user_id": request.user_id,
            "request_message": request.message,
            "ai_response": response,
            "model": request.model
        }
    
    except Exception as e:
        logger.error(f"Error processing text conversation: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/conversation/multimodal")
async def create_multimodal_conversation(request: MultimodalConversationRequest):
    """创建多模态对话的端点（文本+图片）"""
    try:
        logger.info(f"Processing multimodal conversation for user: {request.user_id}")
        logger.info(f"Text: {request.text}")
        logger.info(f"Image URL: {request.image_url}")
        
        # 使用Doubao客户端进行多模态对话
        response = doubao_client.multimodal_chat(
            text=request.text,
            image_url=request.image_url,
            model=request.model or "doubao-seed-1-6-250615"
        )
        
        return {
            "message": "Multimodal conversation processed successfully",
            "user_id": request.user_id,
            "request_text": request.text,
            "request_image_url": request.image_url,
            "ai_response": response,
            "model": request.model
        }
    
    except Exception as e:
        logger.error(f"Error processing multimodal conversation: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/test/doubao")
async def test_doubao_api():
    """测试Doubao API连接"""
    try:
        # 测试简单文本聊天
        text_response = doubao_client.simple_chat("你好，请简单回复一句话")
        
        # 测试多模态聊天
        multimodal_response = doubao_client.multimodal_chat(
            text="图片主要讲了什么?",
            image_url="https://ark-project.tos-cn-beijing.ivolces.com/images/view.jpeg"
        )
        
        return {
            "status": "success",
            "text_test": {
                "request": "你好，请简单回复一句话",
                "response": text_response
            },
            "multimodal_test": {
                "request": "图片主要讲了什么?",
                "image_url": "https://ark-project.tos-cn-beijing.ivolces.com/images/view.jpeg",
                "response": multimodal_response
            }
        }
    
    except Exception as e:
        logger.error(f"Doubao API test failed: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Doubao API test failed: {str(e)}")

if __name__ == "__main__":
    import uvicorn
    
    # 验证API Key是否正确加载
    try:
        logger.info("Verifying API Key configuration...")
        api_key = config.get_api_key()
        logger.info(f"API Key successfully loaded: {'*' * (len(api_key) - 4) + api_key[-4:]}")
    except Exception as e:
        logger.error(f"Failed to load API Key: {e}")
        exit(1)
    
    # 启动服务
    uvicorn.run(app, host="0.0.0.0", port=8000)
