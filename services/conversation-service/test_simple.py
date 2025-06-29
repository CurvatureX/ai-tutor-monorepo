#!/usr/bin/env python3
"""
Simple test script for the simplified conversation service
"""

import asyncio
import json
import httpx

BASE_URL = "http://localhost:8000"
TEST_TOKEN = "demo-token"


async def test_health():
    """Test health endpoint"""
    async with httpx.AsyncClient() as client:
        response = await client.get(f"{BASE_URL}/health")
        print(f"✓ Health check: {response.status_code}")
        print(f"  Response: {response.json()}")
        return response.status_code == 200


async def test_models():
    """Test models endpoint"""
    headers = {"Authorization": f"Bearer {TEST_TOKEN}"}
    async with httpx.AsyncClient() as client:
        response = await client.get(f"{BASE_URL}/v1/models", headers=headers)
        print(f"✓ Models list: {response.status_code}")
        if response.status_code == 200:
            models = response.json()["data"]
            print(f"  Available models: {[m['id'] for m in models]}")
        return response.status_code == 200


async def test_simple_conversation():
    """Test simple conversation endpoint"""
    headers = {
        "Authorization": f"Bearer {TEST_TOKEN}",
        "Content-Type": "application/json",
    }
    data = {"message": "Hello! I want to practice English."}

    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{BASE_URL}/conversation", headers=headers, json=data
        )
        print(f"✓ Simple conversation: {response.status_code}")
        if response.status_code == 200:
            result = response.json()
            print(f"  Response: {result['response'][:100]}...")
        return response.status_code == 200


async def test_chat_completion():
    """Test OpenAI-compatible chat completion"""
    headers = {
        "Authorization": f"Bearer {TEST_TOKEN}",
        "Content-Type": "application/json",
    }
    data = {
        "model": "gpt-4-tutor",
        "messages": [{"role": "user", "content": "Help me practice English grammar"}],
        "stream": False,
    }

    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{BASE_URL}/v1/chat/completions", headers=headers, json=data
        )
        print(f"✓ Chat completion: {response.status_code}")
        if response.status_code == 200:
            result = response.json()
            content = result["choices"][0]["message"]["content"]
            print(f"  Response: {content[:100]}...")
        return response.status_code == 200


async def test_language_session():
    """Test language session creation"""
    headers = {
        "Authorization": f"Bearer {TEST_TOKEN}",
        "Content-Type": "application/json",
    }

    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{BASE_URL}/v1/language/session?language=English&level=intermediate",
            headers=headers,
        )
        print(f"✓ Language session: {response.status_code}")
        if response.status_code == 200:
            session = response.json()
            print(f"  Session ID: {session['id']}")
            print(f"  Language: {session['language']}, Level: {session['level']}")
        return response.status_code == 200


async def main():
    """Run all tests"""
    print("🧪 Testing Simplified Conversation Service\n")

    tests = [
        ("Health Check", test_health),
        ("Models List", test_models),
        ("Simple Conversation", test_simple_conversation),
        ("Chat Completion", test_chat_completion),
        ("Language Session", test_language_session),
    ]

    results = []
    for test_name, test_func in tests:
        print(f"Running {test_name}...")
        try:
            result = await test_func()
            results.append(result)
            print(
                f"{'✅' if result else '❌'} {test_name}: {'PASSED' if result else 'FAILED'}"
            )
        except Exception as e:
            print(f"❌ {test_name}: ERROR - {e}")
            results.append(False)
        print()

    # Summary
    passed = sum(results)
    total = len(results)
    print(f"📊 Test Summary: {passed}/{total} tests passed")

    if passed == total:
        print("🎉 All tests passed! The simplified service is working correctly.")
    else:
        print(
            "⚠️  Some tests failed. Check the service is running on http://localhost:8000"
        )


if __name__ == "__main__":
    print("Make sure the service is running: python app.py")
    print("Then run this test: python test_simple.py\n")
    asyncio.run(main())
