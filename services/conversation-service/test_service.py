#!/usr/bin/env python3
"""
Test script for AI Tutor Conversation Service
测试 AI 对话服务的脚本
"""

import asyncio
import httpx
import json

BASE_URL = "http://localhost:8000"
DEV_TOKEN = "dev-token"  # 开发模式下使用的测试 token


async def test_service():
    """测试服务的各个功能"""

    headers = {
        "Authorization": f"Bearer {DEV_TOKEN}",
        "Content-Type": "application/json",
    }

    async with httpx.AsyncClient() as client:

        print("🔍 测试服务健康状态...")
        try:
            response = await client.get(f"{BASE_URL}/health")
            print(f"✅ 健康检查: {response.status_code}")
            print(f"   响应: {response.json()}")
        except Exception as e:
            print(f"❌ 健康检查失败: {e}")
            return

        print("\n📋 测试模型列表...")
        try:
            response = await client.get(f"{BASE_URL}/v1/models", headers=headers)
            if response.status_code == 200:
                models = response.json()
                print(f"✅ 可用模型: {len(models['data'])} 个")
                for model in models["data"]:
                    print(f"   - {model['id']}")
            else:
                print(f"❌ 获取模型失败: {response.status_code}")
        except Exception as e:
            print(f"❌ 模型列表测试失败: {e}")

        print("\n🎓 测试创建教学会话...")
        try:
            session_data = {"subject": "数学", "level": "高中"}
            response = await client.post(
                f"{BASE_URL}/v1/tutor/session", headers=headers, params=session_data
            )
            if response.status_code == 200:
                session = response.json()
                print(f"✅ 创建会话成功: {session['id']}")
                session_id = session["id"]
            else:
                print(f"❌ 创建会话失败: {response.status_code}")
                session_id = None
        except Exception as e:
            print(f"❌ 创建会话测试失败: {e}")
            session_id = None

        print("\n💬 测试聊天完成...")
        try:
            chat_request = {
                "model": "gpt-4-tutor",
                "messages": [
                    {"role": "user", "content": "请用简单的语言解释什么是二次方程"}
                ],
                "temperature": 0.7,
                "max_tokens": 500,
            }

            response = await client.post(
                f"{BASE_URL}/v1/chat/completions", headers=headers, json=chat_request
            )

            if response.status_code == 200:
                completion = response.json()
                print("✅ 聊天完成成功")
                print(f"   模型: {completion['model']}")
                print(
                    f"   消息: {completion['choices'][0]['message']['content'][:100]}..."
                )
                print(f"   Token 使用: {completion['usage']}")
            else:
                print(f"❌ 聊天完成失败: {response.status_code}")
                print(f"   错误: {response.text}")
        except Exception as e:
            print(f"❌ 聊天测试失败: {e}")

        if session_id:
            print(f"\n📚 测试获取会话列表...")
            try:
                response = await client.get(
                    f"{BASE_URL}/v1/tutor/sessions", headers=headers
                )
                if response.status_code == 200:
                    sessions = response.json()
                    print(f"✅ 获取会话成功: {len(sessions['sessions'])} 个会话")
                else:
                    print(f"❌ 获取会话失败: {response.status_code}")
            except Exception as e:
                print(f"❌ 获取会话测试失败: {e}")


def main():
    """主函数"""
    print("🚀 开始测试 AI Tutor Conversation Service")
    print("=" * 50)

    try:
        asyncio.run(test_service())
    except KeyboardInterrupt:
        print("\n\n⏹️  测试被用户中断")
    except Exception as e:
        print(f"\n\n❌ 测试过程中发生错误: {e}")

    print("\n" + "=" * 50)
    print("🏁 测试完成")


if __name__ == "__main__":
    main()
