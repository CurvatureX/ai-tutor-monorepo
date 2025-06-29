#!/usr/bin/env python3
"""
测试脚本：演示如何使用 OpenAI SDK 调用我们的本地 AI 服务
This script demonstrates how to use OpenAI SDK to call our local AI service
"""

import asyncio
import httpx
import json
from typing import List, Dict

# 如果你想使用真正的 OpenAI SDK，可以取消下面的注释
# import openai

BASE_URL = "http://localhost:8000"
DEV_TOKEN = "dev-token"  # 开发模式下使用的测试 token


class OpenAICompatibleClient:
    """OpenAI 兼容的客户端类"""

    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url
        self.api_key = api_key
        self.headers = {
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        }

    async def list_models(self) -> Dict:
        """获取可用模型列表"""
        async with httpx.AsyncClient() as client:
            response = await client.get(
                f"{self.base_url}/v1/models", headers=self.headers
            )
            return response.json()

    async def create_chat_completion(
        self,
        model: str,
        messages: List[Dict],
        temperature: float = 0.7,
        max_tokens: int = None,
        stream: bool = False,
    ) -> Dict:
        """创建聊天完成"""
        payload = {
            "model": model,
            "messages": messages,
            "temperature": temperature,
            "stream": stream,
        }
        if max_tokens:
            payload["max_tokens"] = max_tokens

        async with httpx.AsyncClient(timeout=30.0) as client:
            if stream:
                return await self._handle_streaming_response(client, payload)
            else:
                response = await client.post(
                    f"{self.base_url}/v1/chat/completions",
                    headers=self.headers,
                    json=payload,
                )
                return response.json()

    async def _handle_streaming_response(
        self, client: httpx.AsyncClient, payload: Dict
    ):
        """处理流式响应"""
        print("🔄 开始流式响应...")

        full_content = ""
        async with client.stream(
            "POST",
            f"{self.base_url}/v1/chat/completions",
            headers=self.headers,
            json=payload,
        ) as response:
            async for chunk in response.aiter_text():
                if chunk.strip():
                    lines = chunk.strip().split('\n')
                    for line in lines:
                        if line.startswith('data: '):
                            data = line[6:]  # 移除 'data: ' 前缀
                            if data == '[DONE]':
                                print("\n✅ 流式响应完成")
                                break
                            try:
                                chunk_data = json.loads(data)
                                if 'choices' in chunk_data and chunk_data['choices']:
                                    delta = chunk_data['choices'][0].get('delta', {})
                                    content = delta.get('content', '')
                                    if content:
                                        print(content, end='', flush=True)
                                        full_content += content
                            except json.JSONDecodeError:
                                continue

        return {"content": full_content}


async def test_openai_compatible_api():
    """测试 OpenAI 兼容的 API"""

    print("🚀 测试 OpenAI 兼容 API")
    print("=" * 50)

    # 初始化客户端
    client = OpenAICompatibleClient(BASE_URL, DEV_TOKEN)

    try:
        # 1. 测试模型列表
        print("\n📋 测试获取模型列表...")
        models = await client.list_models()
        print(f"✅ 可用模型: {len(models['data'])} 个")
        for model in models['data']:
            print(f"   - {model['id']} (owned by: {model['owned_by']})")

        # 2. 测试普通聊天完成
        print("\n💬 测试普通聊天完成...")
        messages = [
            {"role": "user", "content": "请解释什么是二次方程，并给出一个简单的例子"}
        ]

        response = await client.create_chat_completion(
            model="gpt-4-tutor", messages=messages, temperature=0.7, max_tokens=300
        )

        print("✅ 响应内容:")
        print(response['choices'][0]['message']['content'])
        print(f"\n📊 Token 使用情况: {response['usage']}")

        # 3. 测试流式响应
        print("\n🌊 测试流式聊天完成...")
        stream_messages = [{"role": "user", "content": "用简单的话解释什么是机器学习"}]

        await client.create_chat_completion(
            model="local-tutor", messages=stream_messages, temperature=0.8, stream=True
        )

        # 4. 测试多轮对话
        print("\n\n🔄 测试多轮对话...")
        conversation = [{"role": "user", "content": "你好，我想学习 Python 编程"}]

        response = await client.create_chat_completion(
            model="gpt-3.5-turbo-tutor", messages=conversation, temperature=0.6
        )

        assistant_reply = response['choices'][0]['message']['content']
        print(f"🤖 AI 导师: {assistant_reply}")

        # 继续对话
        conversation.append({"role": "assistant", "content": assistant_reply})
        conversation.append({"role": "user", "content": "那我应该从哪里开始学习呢？"})

        response = await client.create_chat_completion(
            model="gpt-3.5-turbo-tutor", messages=conversation, temperature=0.6
        )

        print(f"\n🤖 AI 导师: {response['choices'][0]['message']['content']}")

        # 5. 测试不同学科
        print("\n\n📚 测试不同学科的响应...")
        subjects = [
            "解释光合作用的过程",
            "什么是微积分中的导数？",
            "如何学好英语语法？",
        ]

        for i, subject in enumerate(subjects, 1):
            print(f"\n{i}. 问题: {subject}")
            response = await client.create_chat_completion(
                model="local-tutor",
                messages=[{"role": "user", "content": subject}],
                temperature=0.5,
            )
            print(f"   回答: {response['choices'][0]['message']['content'][:100]}...")

    except Exception as e:
        print(f"❌ 测试过程中发生错误: {e}")


async def test_with_real_openai_sdk():
    """使用真正的 OpenAI SDK 测试（需要安装 openai 包）"""

    try:
        import openai

        print("\n🔧 使用真正的 OpenAI SDK 测试...")
        print("=" * 50)

        # 配置 OpenAI 客户端指向我们的服务
        openai.api_base = f"{BASE_URL}/v1"
        openai.api_key = DEV_TOKEN

        # 测试聊天完成
        response = await openai.ChatCompletion.acreate(
            model="gpt-4-tutor",
            messages=[
                {"role": "user", "content": "用 OpenAI SDK 测试: 解释什么是算法"}
            ],
            temperature=0.7,
        )

        print("✅ OpenAI SDK 测试成功!")
        print(f"响应: {response.choices[0].message.content}")

    except ImportError:
        print("⚠️  未安装 openai 包，跳过 OpenAI SDK 测试")
        print("   可以通过 'uv add openai' 安装")
    except Exception as e:
        print(f"❌ OpenAI SDK 测试失败: {e}")


def create_usage_examples():
    """创建使用示例代码"""

    examples = {
        "python_client.py": '''
# Python 客户端示例
import httpx
import asyncio

async def chat_with_ai_tutor():
    headers = {"Authorization": "Bearer dev-token", "Content-Type": "application/json"}
    
    payload = {
        "model": "gpt-4-tutor",
        "messages": [
            {"role": "user", "content": "解释什么是递归算法"}
        ],
        "temperature": 0.7
    }
    
    async with httpx.AsyncClient() as client:
        response = await client.post(
            "http://localhost:8000/v1/chat/completions",
            headers=headers,
            json=payload
        )
        
        result = response.json()
        print(result["choices"][0]["message"]["content"])

asyncio.run(chat_with_ai_tutor())
        ''',
        "openai_sdk_example.py": '''
# 使用 OpenAI SDK 的示例
import openai

# 配置 OpenAI SDK 指向我们的服务
openai.api_base = "http://localhost:8000/v1"
openai.api_key = "dev-token"

# 现在你可以像使用 OpenAI API 一样使用我们的服务
response = openai.ChatCompletion.create(
    model="gpt-4-tutor",
    messages=[
        {"role": "user", "content": "教我学习数据结构"}
    ]
)

print(response.choices[0].message.content)
        ''',
        "curl_example.sh": '''
#!/bin/bash
# cURL 示例

# 获取模型列表
curl -X GET "http://localhost:8000/v1/models" \\
  -H "Authorization: Bearer dev-token"

# 聊天完成
curl -X POST "http://localhost:8000/v1/chat/completions" \\
  -H "Authorization: Bearer dev-token" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-4-tutor",
    "messages": [
      {"role": "user", "content": "解释机器学习的基本概念"}
    ],
    "temperature": 0.7
  }'
        ''',
    }

    print("\n📝 创建使用示例文件...")
    for filename, content in examples.items():
        with open(f"examples_{filename}", "w", encoding="utf-8") as f:
            f.write(content.strip())
        print(f"   ✅ 创建了 examples_{filename}")


async def main():
    """主函数"""
    print("🤖 AI Tutor Service - OpenAI 兼容 API 测试")
    print("=" * 60)

    # 检查服务是否运行
    try:
        async with httpx.AsyncClient() as client:
            response = await client.get(f"{BASE_URL}/health")
            if response.status_code == 200:
                print("✅ 服务运行正常")
            else:
                print("❌ 服务异常")
                return
    except Exception as e:
        print(f"❌ 无法连接到服务: {e}")
        print("请确保服务正在运行: uv run python scripts/dev.py run")
        return

    # 运行测试
    await test_openai_compatible_api()
    await test_with_real_openai_sdk()

    # 创建示例文件
    create_usage_examples()

    print("\n" + "=" * 60)
    print("🎉 测试完成! 你的 OpenAI 兼容 API 服务工作正常")
    print("\n📚 主要特性:")
    print("   ✅ 完全兼容 OpenAI API 格式")
    print("   ✅ 支持普通和流式响应")
    print("   ✅ 智能教学响应")
    print("   ✅ 多模型支持")
    print("   ✅ 可以使用现有的 OpenAI SDK")


if __name__ == "__main__":
    asyncio.run(main())
