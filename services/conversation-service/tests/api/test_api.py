#!/usr/bin/env python3
"""
测试Conversation Service API的脚本
"""

import requests
import json
import time

BASE_URL = "http://localhost:8000"

def test_health_check():
    """测试健康检查端点"""
    print("🔍 Testing health check...")
    try:
        response = requests.get(f"{BASE_URL}/health")
        response.raise_for_status()
        data = response.json()
        print(f"✅ Health check passed: {data}")
        return True
    except requests.exceptions.RequestException as e:
        print(f"❌ Health check failed: {e}")
        return False

def test_text_conversation():
    """测试文本对话"""
    print("\n💬 Testing text conversation...")
    try:
        payload = {
            "message": "请简单介绍一下人工智能",
            "user_id": "test_user_123"
        }
        
        response = requests.post(
            f"{BASE_URL}/conversation",
            json=payload,
            headers={"Content-Type": "application/json"}
        )
        response.raise_for_status()
        data = response.json()
        
        print(f"✅ Text conversation successful:")
        print(f"   Request: {data.get('request_message')}")
        print(f"   Response: {data.get('ai_response')[:100]}...")
        return True
        
    except requests.exceptions.RequestException as e:
        print(f"❌ Text conversation failed: {e}")
        if hasattr(e, 'response') and e.response is not None:
            print(f"   Response content: {e.response.text}")
        return False

def test_multimodal_conversation():
    """测试多模态对话"""
    print("\n🖼️ Testing multimodal conversation...")
    try:
        payload = {
            "text": "请详细描述这张图片",
            "image_url": "https://ark-project.tos-cn-beijing.ivolces.com/images/view.jpeg",
            "user_id": "test_user_123"
        }
        
        response = requests.post(
            f"{BASE_URL}/conversation/multimodal",
            json=payload,
            headers={"Content-Type": "application/json"}
        )
        response.raise_for_status()
        data = response.json()
        
        print(f"✅ Multimodal conversation successful:")
        print(f"   Request: {data.get('request_text')}")
        print(f"   Image: {data.get('request_image_url')}")
        print(f"   Response: {data.get('ai_response')[:150]}...")
        return True
        
    except requests.exceptions.RequestException as e:
        print(f"❌ Multimodal conversation failed: {e}")
        if hasattr(e, 'response') and e.response is not None:
            print(f"   Response content: {e.response.text}")
        return False

def test_doubao_api():
    """测试Doubao API"""
    print("\n🧪 Testing Doubao API directly...")
    try:
        response = requests.post(f"{BASE_URL}/test/doubao")
        response.raise_for_status()
        data = response.json()
        
        print(f"✅ Doubao API test successful:")
        print(f"   Text test: {data.get('text_test', {}).get('response', 'N/A')[:50]}...")
        print(f"   Multimodal test: {data.get('multimodal_test', {}).get('response', 'N/A')[:50]}...")
        return True
        
    except requests.exceptions.RequestException as e:
        print(f"❌ Doubao API test failed: {e}")
        if hasattr(e, 'response') and e.response is not None:
            print(f"   Response content: {e.response.text}")
        return False

def wait_for_server(max_attempts=30, delay=1):
    """等待服务器启动"""
    print(f"⏳ Waiting for server to start (max {max_attempts}s)...")
    
    for attempt in range(max_attempts):
        try:
            response = requests.get(f"{BASE_URL}/health", timeout=2)
            if response.status_code == 200:
                print("✅ Server is ready!")
                return True
        except requests.exceptions.RequestException:
            pass
        
        time.sleep(delay)
        print(f"   Attempt {attempt + 1}/{max_attempts}...")
    
    print("❌ Server failed to start within the timeout period")
    return False

def main():
    """主测试函数"""
    print("🚀 Starting Conversation Service API Tests")
    print("=" * 50)
    
    # 等待服务器启动
    if not wait_for_server():
        print("❌ Cannot proceed with tests - server is not responding")
        return
    
    # 运行测试
    tests = [
        ("Health Check", test_health_check),
        ("Text Conversation", test_text_conversation),
        ("Multimodal Conversation", test_multimodal_conversation),
        ("Doubao API Test", test_doubao_api)
    ]
    
    results = []
    for test_name, test_func in tests:
        success = test_func()
        results.append((test_name, success))
    
    # 显示结果
    print("\n" + "=" * 50)
    print("📊 Test Results:")
    for test_name, success in results:
        status = "✅ PASSED" if success else "❌ FAILED"
        print(f"   {test_name}: {status}")
    
    total_tests = len(results)
    passed_tests = sum(1 for _, success in results if success)
    print(f"\n🎯 Summary: {passed_tests}/{total_tests} tests passed")
    
    if passed_tests == total_tests:
        print("🎉 All tests passed! The Conversation Service is working correctly.")
    else:
        print("⚠️  Some tests failed. Please check the logs above.")

if __name__ == "__main__":
    main() 