import os
import requests
import json
import logging
from typing import List, Dict, Any, Optional
from dotenv import load_dotenv

# 配置日志
logger = logging.getLogger(__name__)

class DoubaoClient:
    """Doubao API客户端"""
    
    def __init__(self, api_key: Optional[str] = None):
        """初始化Doubao客户端
        
        Args:
            api_key: API密钥，如果不提供则从环境变量读取
        """
        # 加载环境变量
        load_dotenv()
        
        self.api_key = api_key or os.getenv("DOUBAO_API_KEY")
        if not self.api_key:
            raise ValueError("API key is required. Please set DOUBAO_API_KEY in .env file or pass it directly.")
        
        self.base_url = "https://ark.cn-beijing.volces.com/api/v3"
        self.headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self.api_key}"
        }
    
    def chat_completions(
        self, 
        messages: List[Dict[str, Any]], 
        model: str = "doubao-seed-1-6-250615",
        **kwargs
    ) -> Dict[str, Any]:
        """调用Doubao聊天完成API
        
        Args:
            messages: 消息列表
            model: 使用的模型名称
            **kwargs: 其他API参数
            
        Returns:
            API响应结果
        """
        url = f"{self.base_url}/chat/completions"
        
        payload = {
            "model": model,
            "messages": messages,
            **kwargs
        }
        
        try:
            logger.info(f"Calling Doubao API with model: {model}")
            logger.debug(f"Request payload: {json.dumps(payload, indent=2, ensure_ascii=False)}")
            
            response = requests.post(
                url=url,
                headers=self.headers,
                json=payload,
                timeout=30
            )
            
            response.raise_for_status()
            result = response.json()
            
            logger.info("Doubao API call successful")
            logger.debug(f"Response: {json.dumps(result, indent=2, ensure_ascii=False)}")
            
            return result
            
        except requests.exceptions.RequestException as e:
            logger.error(f"Doubao API request failed: {str(e)}")
            if hasattr(e, 'response') and e.response is not None:
                logger.error(f"Response status: {e.response.status_code}")
                logger.error(f"Response content: {e.response.text}")
            raise
        except Exception as e:
            logger.error(f"Unexpected error in Doubao API call: {str(e)}")
            raise
    
    def create_text_message(self, content: str, role: str = "user") -> Dict[str, Any]:
        """创建文本消息
        
        Args:
            content: 消息内容
            role: 消息角色 (user, assistant, system)
            
        Returns:
            格式化的消息字典
        """
        return {
            "role": role,
            "content": content
        }
    
    def create_multimodal_message(
        self, 
        text: str, 
        image_url: str, 
        role: str = "user"
    ) -> Dict[str, Any]:
        """创建多模态消息（文本+图片）
        
        Args:
            text: 文本内容
            image_url: 图片URL
            role: 消息角色
            
        Returns:
            格式化的多模态消息字典
        """
        return {
            "role": role,
            "content": [
                {
                    "type": "text",
                    "text": text
                },
                {
                    "type": "image_url",
                    "image_url": {
                        "url": image_url
                    }
                }
            ]
        }
    
    def simple_chat(self, message: str, model: str = "doubao-seed-1-6-250615") -> str:
        """简单的文本聊天
        
        Args:
            message: 用户消息
            model: 使用的模型
            
        Returns:
            AI的回复文本
        """
        messages = [self.create_text_message(message)]
        
        response = self.chat_completions(messages=messages, model=model)
        
        # 提取回复内容
        if "choices" in response and len(response["choices"]) > 0:
            return response["choices"][0]["message"]["content"]
        else:
            raise ValueError("Invalid response format from Doubao API")
    
    def multimodal_chat(
        self, 
        text: str, 
        image_url: str, 
        model: str = "doubao-seed-1-6-250615"
    ) -> str:
        """多模态聊天（文本+图片）
        
        Args:
            text: 用户文本消息
            image_url: 图片URL
            model: 使用的模型
            
        Returns:
            AI的回复文本
        """
        messages = [self.create_multimodal_message(text, image_url)]
        
        response = self.chat_completions(messages=messages, model=model)
        
        # 提取回复内容
        if "choices" in response and len(response["choices"]) > 0:
            return response["choices"][0]["message"]["content"]
        else:
            raise ValueError("Invalid response format from Doubao API")


# 测试函数
def test_doubao_client():
    """测试Doubao客户端功能"""
    try:
        # 初始化客户端
        client = DoubaoClient()
        
        # 测试简单文本聊天
        print("Testing simple text chat...")
        response = client.simple_chat("你好，请简单介绍一下自己")
        print(f"Response: {response}")
        
        # 测试多模态聊天（使用示例图片）
        print("\nTesting multimodal chat...")
        response = client.multimodal_chat(
            text="图片主要讲了什么?",
            image_url="https://ark-project.tos-cn-beijing.ivolces.com/images/view.jpeg"
        )
        print(f"Response: {response}")
        
    except Exception as e:
        print(f"Test failed: {e}")


if __name__ == "__main__":
    # 设置日志级别
    logging.basicConfig(level=logging.INFO)
    
    # 运行测试
    test_doubao_client() 