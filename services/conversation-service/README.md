# AI Tutor Conversation Service

🤖 **OpenAI-compatible API service** that provides AI tutoring conversations using **local models** instead of calling OpenAI API. Built with FastAPI and managed by uv.

## 🎯 核心特性

- **🔄 OpenAI API 完全兼容**: 现有使用 OpenAI SDK 的代码可以直接切换到我们的服务
- **🏠 本地模型支持**: 不依赖外部 API，可以使用自己的 AI 模型
- **🎓 智能教学响应**: 专门为教育场景优化的 AI 响应
- **⚡ uv 管理**: 使用最现代的 Python 包管理工具

## ⚡ Quick Start with uv

### 1. Install uv (if not already installed)

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

### 2. Setup and Run

```bash
# Clone and navigate to the service directory
cd services/conversation-service

# Install all dependencies (including dev dependencies)
uv sync

# Run the service
uv run python main.py

# Or use the development script
uv run python scripts/dev.py run
```

### 3. Test the Service

```bash
# Visit API documentation
open http://localhost:8000/docs

# Test health endpoint
curl http://localhost:8000/health

# Or run the comprehensive test
uv run python test_openai_client.py
```

## 🔌 OpenAI SDK 兼容性

现有的 OpenAI 代码可以直接切换到我们的服务：

```python
import openai

# 只需要改变这两行配置
openai.api_base = "http://localhost:8000/v1"  # 指向我们的服务
openai.api_key = "dev-token"                  # 使用我们的认证

# 其余代码完全不变！
response = openai.ChatCompletion.create(
    model="gpt-4-tutor",
    messages=[
        {"role": "user", "content": "解释机器学习的基本概念"}
    ]
)

print(response.choices[0].message.content)
```

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

# Utilities
uv run python scripts/dev.py clean  # Clean cache
```

## 📊 API 兼容性

### ✅ 支持的 OpenAI API 端点

| 端点                        | 状态        | 说明                 |
| --------------------------- | ----------- | -------------------- |
| `GET /v1/models`            | ✅ 完全兼容 | 列出可用模型         |
| `POST /v1/chat/completions` | ✅ 完全兼容 | 聊天完成（支持流式） |

### 🎓 教学专用端点

| 端点                          | 功能             |
| ----------------------------- | ---------------- |
| `POST /v1/tutor/session`      | 创建教学会话     |
| `GET /v1/tutor/sessions`      | 获取用户会话列表 |
| `GET /v1/tutor/sessions/{id}` | 获取会话历史     |

## 🤖 智能教学功能

我们的 AI 具有专门的教学能力：

- **🧠 苏格拉底式教学**: 通过问题引导学生思考
- **📚 多学科支持**: 数学、科学、编程、语言等
- **🎯 自适应难度**: 根据学生水平调整解释深度
- **💡 举例说明**: 使用生活实例帮助理解
- **🔄 互动式学习**: 鼓励学生参与和提问

## 🔧 配置选项

### Environment Variables

| Variable           | Description           | Default                    |
| ------------------ | --------------------- | -------------------------- |
| `PORT`             | Server port           | `8000`                     |
| `DEBUG`            | Debug mode            | `false`                    |
| `USE_LOCAL_MODEL`  | Enable local AI model | `false`                    |
| `MODEL_PATH`       | Path to local model   | `None`                     |
| `MODEL_NAME`       | Model identifier      | `local-tutor`              |
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

### 1. 基础聊天

```bash
curl -X POST "http://localhost:8000/v1/chat/completions" \
  -H "Authorization: Bearer dev-token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4-tutor",
    "messages": [{"role": "user", "content": "解释什么是递归"}],
    "temperature": 0.7
  }'
```

### 2. 流式响应

```python
import httpx

async def stream_chat():
    async with httpx.AsyncClient() as client:
        async with client.stream(
            "POST", "http://localhost:8000/v1/chat/completions",
            headers={"Authorization": "Bearer dev-token"},
            json={
                "model": "local-tutor",
                "messages": [{"role": "user", "content": "解释机器学习"}],
                "stream": True
            }
        ) as response:
            async for chunk in response.aiter_text():
                print(chunk, end="")
```

### 3. Python 客户端

```python
import httpx
import asyncio

class AITutorClient:
    def __init__(self, base_url="http://localhost:8000", api_key="dev-token"):
        self.base_url = base_url
        self.headers = {"Authorization": f"Bearer {api_key}"}

    async def chat(self, message: str, model="gpt-4-tutor"):
        async with httpx.AsyncClient() as client:
            response = await client.post(
                f"{self.base_url}/v1/chat/completions",
                headers=self.headers,
                json={
                    "model": model,
                    "messages": [{"role": "user", "content": message}]
                }
            )
            return response.json()["choices"][0]["message"]["content"]

# 使用示例
client = AITutorClient()
answer = await client.chat("什么是算法复杂度？")
print(answer)
```

## 🐳 Docker 部署

```bash
# 构建镜像
docker build -t ai-tutor-service .

# 运行容器
docker run -p 8000:8000 \
  -e USE_LOCAL_MODEL=false \
  -e DEBUG=false \
  ai-tutor-service
```

## 📈 性能和扩展

- **⚡ 高性能**: FastAPI + async/await
- **🔄 可扩展**: 支持负载均衡和水平扩展
- **📊 监控**: 内置健康检查和指标
- **🔒 安全**: JWT 认证和 CORS 配置

## 🆚 与 OpenAI API 的对比

| 特性       | OpenAI API    | 我们的服务  |
| ---------- | ------------- | ----------- |
| API 格式   | ✅            | ✅ 完全兼容 |
| 费用       | 💰 按使用付费 | 🆓 完全免费 |
| 数据隐私   | ⚠️ 发送到外部 | 🔒 完全本地 |
| 自定义模型 | ❌ 不支持     | ✅ 完全支持 |
| 教学优化   | ❌ 通用模型   | 🎓 专门优化 |
| 离线使用   | ❌ 需要网络   | ✅ 支持离线 |

## 🔄 迁移指南

### 从 OpenAI API 迁移

1. **无需修改代码**: 只需更改 `api_base` 和 `api_key`
2. **模型映射**: `gpt-4` → `gpt-4-tutor`, `gpt-3.5-turbo` → `gpt-3.5-turbo-tutor`
3. **额外功能**: 可以使用我们的教学专用端点

### 示例迁移

```python
# 原来的代码
import openai
openai.api_key = "sk-..."  # OpenAI 密钥

# 迁移后的代码
import openai
openai.api_base = "http://your-service:8000/v1"  # 你的服务地址
openai.api_key = "your-jwt-token"               # 你的认证令牌

# 其余代码完全不变！
```

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
