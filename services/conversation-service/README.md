# Conversation Service

基于Doubao API的对话服务，支持文本和多模态（文本+图片）对话。

## 功能特性

- ✅ 文本对话：支持基于文本的AI对话
- ✅ 多模态对话：支持文本+图片的AI对话
- ✅ 环境变量配置：从.env文件读取API密钥
- ✅ RESTful API：提供完整的HTTP API接口
- ✅ 健康检查：提供服务状态检查端点
- ✅ 完整测试：包含API功能测试脚本

## 项目结构

```
services/conversation-service/
├── .env                    # 环境变量配置文件
├── main.py                 # FastAPI主应用
├── doubao_client.py        # Doubao API客户端
├── requirements.txt        # Python依赖
├── test_api.py            # API测试脚本
└── README.md              # 项目说明文档
```

## 快速开始

### 1. 安装依赖

```bash
cd services/conversation-service
python3 -m pip install -r requirements.txt
```

### 2. 配置环境变量

`.env`文件已包含Doubao API密钥：

```bash
# Doubao API Configuration
DOUBAO_API_KEY=7cc1663d-7c97-4269-ac45-a49098e1108d
```

### 3. 启动服务

```bash
python3 main.py
```

服务将在 `http://localhost:8000` 启动。

### 4. 测试服务

运行测试脚本：

```bash
# 在另一个终端窗口中运行
python3 test_api.py
```

## API端点

### 健康检查

```http
GET /health
```

响应：
```json
{
  "status": "healthy",
  "service": "conversation-service",
  "api_key_configured": true
}
```

### 文本对话

```http
POST /conversation
Content-Type: application/json

{
  "message": "你好，请介绍一下自己",
  "user_id": "user_123",
  "model": "doubao-seed-1-6-250615"
}
```

响应：
```json
{
  "message": "Text conversation processed successfully",
  "user_id": "user_123",
  "request_message": "你好，请介绍一下自己",
  "ai_response": "你好呀！我叫豆包，是由字节跳动公司开发的人工智能助手...",
  "model": "doubao-seed-1-6-250615"
}
```

### 多模态对话

```http
POST /conversation/multimodal
Content-Type: application/json

{
  "text": "图片主要讲了什么?",
  "image_url": "https://ark-project.tos-cn-beijing.ivolces.com/images/view.jpeg",
  "user_id": "user_123",
  "model": "doubao-seed-1-6-250615"
}
```

响应：
```json
{
  "message": "Multimodal conversation processed successfully",
  "user_id": "user_123",
  "request_text": "图片主要讲了什么?",
  "request_image_url": "https://ark-project.tos-cn-beijing.ivolces.com/images/view.jpeg",
  "ai_response": "图片展现了一幅宁静壮丽的自然景观...",
  "model": "doubao-seed-1-6-250615"
}
```

### Doubao API测试

```http
POST /test/doubao
```

用于测试Doubao API连接状态。

## 使用示例

### Python客户端示例

```python
import requests

# 文本对话
response = requests.post(
    "http://localhost:8000/conversation",
    json={
        "message": "解释一下机器学习",
        "user_id": "test_user"
    }
)
print(response.json()["ai_response"])

# 多模态对话
response = requests.post(
    "http://localhost:8000/conversation/multimodal",
    json={
        "text": "描述这张图片",
        "image_url": "your_image_url_here",
        "user_id": "test_user"
    }
)
print(response.json()["ai_response"])
```

### curl示例

```bash
# 文本对话
curl -X POST "http://localhost:8000/conversation" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "你好，世界",
    "user_id": "test_user"
  }'

# 多模态对话
curl -X POST "http://localhost:8000/conversation/multimodal" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "图片内容是什么？",
    "image_url": "https://example.com/image.jpg",
    "user_id": "test_user"
  }'
```

## 技术实现

### DoubaoClient类

`doubao_client.py`提供了完整的Doubao API封装：

- `chat_completions()`: 原始API调用方法
- `simple_chat()`: 简化的文本对话方法
- `multimodal_chat()`: 多模态对话方法
- `create_text_message()`: 创建文本消息
- `create_multimodal_message()`: 创建多模态消息

### 配置管理

- 使用`python-dotenv`从`.env`文件加载环境变量
- 支持运行时API key验证
- 提供配置状态检查

### 错误处理

- 完整的HTTP错误处理
- 详细的日志记录
- API调用超时保护

## 开发和调试

启动开发模式：

```bash
# 启用详细日志
python3 -c "
import logging
logging.basicConfig(level=logging.DEBUG)
exec(open('main.py').read())
"
```

## 注意事项

1. 确保网络连接正常，能够访问Doubao API
2. API key需要有效并具有相应权限
3. 图片URL需要公开可访问
4. 建议在生产环境中使用HTTPS
5. 注意API调用频率限制

## 故障排除

### 常见问题

1. **API Key错误**：检查`.env`文件中的`DOUBAO_API_KEY`是否正确
2. **网络连接失败**：确认能够访问`https://ark.cn-beijing.volces.com`
3. **图片无法加载**：确认图片URL是否有效且公开可访问
4. **依赖缺失**：运行`pip install -r requirements.txt`安装依赖

### 调试步骤

1. 检查服务启动日志
2. 运行`python3 test_api.py`测试API
3. 使用`python3 doubao_client.py`直接测试Doubao连接
4. 检查`.env`文件配置

## 许可证

本项目遵循相应的开源许可证。 