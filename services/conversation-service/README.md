# AI Tutor Conversation Service

🤖 **Unified API service** that provides AI tutoring conversations supporting both **OpenAI-compatible API** and **Doubao API**. Built with FastAPI and managed by uv.

## 🎯 核心特性

- **🔄 OpenAI API 完全兼容**: 现有使用 OpenAI SDK 的代码可以直接切换到我们的服务
- **🏠 本地模型支持**: 不依赖外部 API，可以使用自己的 AI 模型
- **🎯 Doubao API 集成**: 支持豆包 API 进行文本和多模态对话
- **🎓 智能教学响应**: 专门为教育场景优化的 AI 响应
- **⚡ uv 管理**: 使用最现代的 Python 包管理工具

## ⚡ Quick Start with uv

### 1. Install uv (if not already installed)

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

### 2. Setup Environment

```bash
# Clone and navigate to the service directory
cd services/conversation-service

# Copy environment template
cp .env.example .env

# Edit .env and set your API keys:
# DOUBAO_API_KEY=your_doubao_api_key_here
# JWT_SECRET=your_jwt_secret_here
```

### 3. Install and Run

```bash
# Install all dependencies (including dev dependencies)
uv sync

# Run the service
uv run python main.py

# Or use the development script
uv run python scripts/dev.py run
```

### 4. Test the Service

```bash
# Visit API documentation
open http://localhost:8000/docs

# Test health endpoint
curl http://localhost:8000/health

# Test OpenAI compatibility
uv run python test_openai_client.py

# Test Doubao API
uv run python test_api.py
```

## 🔌 API 兼容性

### ✅ OpenAI-Compatible API

现有的 OpenAI 代码可以直接切换到我们的服务：

```python
import openai

# 只需要改变这两行配置
openai.api_base = "http://localhost:8000/v1"  # 指向我们的服务
openai.api_key = "dev-token"                  # 使用我们的认证

# 其余代码完全不变！
response = openai.ChatCompletion.create(
    model="gpt-4-tutor",  # 或者 "doubao-seed-1-6-250615"
    messages=[
        {"role": "user", "content": "解释机器学习的基本概念"}
    ]
)

print(response.choices[0].message.content)
```

### 🎯 Doubao API Endpoints

```bash
# Text conversation
curl -X POST "http://localhost:8000/conversation" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "你好，我想学习编程",
    "user_id": "user123",
    "model": "doubao-seed-1-6-250615"
  }'

# Multimodal conversation
curl -X POST "http://localhost:8000/conversation/multimodal" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "这张图片讲了什么？",
    "image_url": "https://example.com/image.jpg",
    "user_id": "user123",
    "model": "doubao-seed-1-6-250615"
  }'
```

## 📊 API 端点总览

### OpenAI-Compatible Endpoints

| 端点                      | 方法 | 说明                 | 认证 |
| ------------------------- | ---- | -------------------- | ---- |
| `/v1/models`              | GET  | 列出可用模型         | JWT  |
| `/v1/chat/completions`    | POST | 聊天完成（支持流式） | JWT  |
| `/v1/tutor/session`       | POST | 创建教学会话         | JWT  |
| `/v1/tutor/sessions`      | GET  | 获取用户会话列表     | JWT  |
| `/v1/tutor/sessions/{id}` | GET  | 获取会话历史         | JWT  |

### Doubao API Endpoints

| 端点                       | 方法 | 说明                 | 认证 |
| -------------------------- | ---- | -------------------- | ---- |
| `/conversation`            | POST | 文本对话             | None |
| `/conversation/multimodal` | POST | 多模态对话           | None |
| `/test/doubao`             | POST | 测试 Doubao API 连接 | None |

### 通用端点

| 端点      | 方法 | 说明     |
| --------- | ---- | -------- |
| `/`       | GET  | 服务信息 |
| `/health` | GET  | 健康检查 |
| `/docs`   | GET  | API 文档 |

## 🤖 支持的模型

| 模型 ID                  | 提供商 | 功能     | API    |
| ------------------------ | ------ | -------- | ------ |
| `gpt-4-tutor`            | Local  | 智能教学 | OpenAI |
| `gpt-3.5-turbo-tutor`    | Local  | 智能教学 | OpenAI |
| `local-tutor`            | Local  | 智能教学 | OpenAI |
| `doubao-seed-1-6-250615` | Doubao | 通用对话 | Both   |

## 🛠️ Development Commands

```bash
# Install dependencies
uv sync                              # Production only
uv run python scripts/dev.py dev    # All dependencies

# Run service
uv run python scripts/dev.py run

# Code quality
uv run python scripts/dev.py lint   # Check code
uv run python scripts/dev.py format # Format code

# Testing
uv run python scripts/dev.py test   # Run tests
uv run python test_openai_client.py # Test OpenAI compatibility
uv run python test_api.py           # Test Doubao API

# Utilities
uv run python scripts/dev.py clean  # Clean cache
```

## 🔧 配置选项

### Environment Variables

| Variable           | Description           | Default                    |
| ------------------ | --------------------- | -------------------------- |
| `PORT`             | Server port           | `8000`                     |
| `DEBUG`            | Debug mode            | `false`                    |
| `USE_LOCAL_MODEL`  | Enable local AI model | `false`                    |
| `MODEL_PATH`       | Path to local model   | `None`                     |
| `MODEL_NAME`       | Model identifier      | `local-tutor`              |
| `DOUBAO_API_KEY`   | Doubao API key        | Required                   |
| `JWT_SECRET`       | JWT signing secret    | `your-secret-key`          |
| `USER_SERVICE_URL` | User service URL      | `http://user-service:8001` |
| `REDIS_URL`        | Redis connection URL  | `redis://localhost:6379`   |

### 本地模型集成

```python
# 在 ai_service.py 中启用真实模型
async def _load_local_model(self):
    from transformers import AutoTokenizer, AutoModelForCausalLM

    self.tokenizer = AutoTokenizer.from_pretrained("your-model-name")
    self.local_model = AutoModelForCausalLM.from_pretrained("your-model-name")
```

## 🚀 使用示例

### 1. OpenAI 兼容调用

```python
import httpx

async def chat_with_openai_format():
    headers = {"Authorization": "Bearer dev-token", "Content-Type": "application/json"}

    payload = {
        "model": "gpt-4-tutor",
        "messages": [
            {"role": "user", "content": "什么是算法复杂度？"}
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
```

### 2. Doubao API 调用

```python
import httpx

async def chat_with_doubao():
    payload = {
        "message": "解释什么是深度学习",
        "user_id": "user123",
        "model": "doubao-seed-1-6-250615"
    }

    async with httpx.AsyncClient() as client:
        response = await client.post(
            "http://localhost:8000/conversation",
            json=payload
        )
        result = response.json()
        print(result["response"])
```

### 3. 多模态对话

```python
import httpx

async def multimodal_chat():
    payload = {
        "text": "这张图片展示了什么内容？",
        "image_url": "https://example.com/image.jpg",
        "user_id": "user123",
        "model": "doubao-seed-1-6-250615"
    }

    async with httpx.AsyncClient() as client:
        response = await client.post(
            "http://localhost:8000/conversation/multimodal",
            json=payload
        )
        result = response.json()
        print(result["response"])
```

## 🐳 Docker 部署

```bash
# 构建镜像
docker build -t ai-tutor-service .

# 运行容器
docker run -p 8000:8000 \
  -e USE_LOCAL_MODEL=false \
  -e DOUBAO_API_KEY=your_api_key \
  -e DEBUG=false \
  ai-tutor-service
```

## 📈 性能和扩展

- **⚡ 高性能**: FastAPI + async/await
- **🔄 可扩展**: 支持负载均衡和水平扩展
- **📊 监控**: 内置健康检查和指标
- **🔒 安全**: JWT 认证和 CORS 配置
- **🎛️ 灵活**: 支持多种 AI 后端

## 🆚 API 对比

| 特性       | OpenAI API    | Doubao API    | 我们的服务          |
| ---------- | ------------- | ------------- | ------------------- |
| 格式标准   | ✅ OpenAI     | ✅ Doubao     | ✅ 两者都支持       |
| 费用       | 💰 按使用付费 | 💰 按使用付费 | 🆓 本地模型免费     |
| 数据隐私   | ⚠️ 发送到外部 | ⚠️ 发送到外部 | 🔒 本地模型完全私有 |
| 自定义模型 | ❌ 不支持     | ❌ 不支持     | ✅ 完全支持         |
| 教学优化   | ❌ 通用模型   | ❌ 通用模型   | 🎓 专门优化         |
| 多模态     | ✅ 支持       | ✅ 支持       | ✅ 支持             |

## 🔄 迁移指南

### 从 OpenAI API 迁移

1. **无需修改代码**: 只需更改 `api_base` 和 `api_key`
2. **模型选择**: 可选择本地模型或 Doubao 模型
3. **额外功能**: 使用教学专用端点

### 从其他服务迁移

1. **使用 Doubao API**: 直接调用 `/conversation` 端点
2. **多模态支持**: 使用 `/conversation/multimodal` 端点
3. **灵活集成**: 选择最适合的 API 格式

## 🤝 贡献

1. Fork 项目
2. 创建功能分支: `git checkout -b feature/amazing-feature`
3. 提交更改: `git commit -m 'Add amazing feature'`
4. 推送分支: `git push origin feature/amazing-feature`
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- FastAPI 团队提供优秀的 Web 框架
- Hugging Face 提供 transformers 库
- OpenAI 提供 API 标准参考
- 豆包（Doubao）提供 AI 服务支持
