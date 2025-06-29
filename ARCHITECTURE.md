# 🏗️ AI Tutor Monorepo 架构文档

## 📋 目录

- [系统概览](#系统概览)
- [架构设计](#架构设计)
- [技术栈](#技术栈)
- [核心组件](#核心组件)
- [数据流设计](#数据流设计)
- [用户交互时序图](#用户交互时序图)
- [API设计](#api设计)
- [部署架构](#部署架构)
- [安全设计](#安全设计)
- [监控与日志](#监控与日志)

## 系统概览

AI Tutor Monorepo 是一个基于微服务架构的AI英语教学平台，提供实时语音对话、智能评分、个性化学习等功能。系统采用现代化的技术栈，支持多平台客户端，具备高可用、高并发、可扩展的特性。

### 🎯 核心功能

- **实时语音对话**: 支持语音识别(ASR)、语言模型(LLM)、语音合成(TTS)
- **智能语音评分**: 基于科大讯飞ISE API的多维度发音评测
- **多平台支持**: Web端、Flutter移动端、Unity 3D沉浸式环境
- **个性化学习**: 用户画像、学习进度跟踪、智能推荐

### 🏛️ 架构原则

- **微服务架构**: 服务独立部署、横向扩展、故障隔离
- **事件驱动**: 异步消息处理、解耦合设计
- **API优先**: RESTful + gRPC 双协议支持
- **云原生**: 容器化部署、Kubernetes编排

## 架构设计

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Applications                       │
├─────────────────┬─────────────────┬─────────────────────────┤
│   Web Frontend  │  Flutter App    │    Unity 3D App         │
│   (React/Vue)   │   (iOS/Android) │  (VR/AR Learning)       │
└─────────────────┴─────────────────┴─────────────────────────┘
                            │
                            │ WebSocket/HTTP/gRPC
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway                            │
│                    (Go + Gin)                               │
│  • Request Routing    • Load Balancing   • Rate Limiting   │
│  • Authentication     • Protocol Convert • Logging         │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ gRPC
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Microservices Layer                       │
├─────────────────┬─────────────────┬─────────────────────────┤
│  Speech Service │  Conversation   │    User Service         │
│  (Go + gRPC)    │   Service       │   (Python + gRPC)      │
│  • ASR (Volcano)│  (Python)       │   • Authentication     │
│  • TTS (Volcano)│  • LLM (Doubao) │   • User Profiles      │
│  • ISE (iFlytek)│  • Context Mgmt │   • Progress Tracking  │
└─────────────────┼─────────────────┼─────────────────────────┤
│ Analytics Service│ Notification   │   Shared Components     │
│ (Python + ML)    │  Service       │   • Proto Definitions  │
│ • Learning Stats │  (Go + gRPC)   │   • Common Utilities    │
│ • Performance    │  • Push Alerts │   • Config Management  │
└─────────────────┴─────────────────┴─────────────────────────┘
                            │
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                       │
├─────────────────┬─────────────────┬─────────────────────────┤
│    Databases    │   Message Queue │    External APIs        │
│  • PostgreSQL   │   • Redis       │  • Volcano Engine       │
│  • MongoDB      │   • RabbitMQ    │  • Doubao LLM           │
│  • Redis Cache  │   • Kafka       │  • iFlytek ISE          │
└─────────────────┴─────────────────┴─────────────────────────┘
```

### 服务间通信架构

```
Frontend (WebSocket) ←→ Gateway (HTTP/WebSocket/gRPC) ←→ Services (gRPC)
                                   ↕
                            Message Queue (Events)
                                   ↕
                            Databases & External APIs
```

## 技术栈

### 前端技术栈

| 组件 | 技术选型 | 说明 |
|------|----------|------|
| Web Frontend | HTML5 + CSS3 + JavaScript | 现代化响应式UI，WebSocket实时通信 |
| Mobile App | Flutter + Dart | 跨平台移动应用开发 |
| 3D Environment | Unity + C# | 沉浸式学习环境，VR/AR支持 |

### 后端技术栈

| 组件 | 技术选型 | 说明 |
|------|----------|------|
| API Gateway | Go + Gin + WebSocket | 高性能网关，协议转换 |
| Speech Service | Go + gRPC + WebSocket | 语音处理核心服务 |
| Conversation Service | Python + FastAPI + gRPC | AI对话服务 |
| User Service | Python + FastAPI + gRPC | 用户管理服务 |
| Analytics Service | Python + Pandas + ML | 数据分析与机器学习 |
| Notification Service | Go + gRPC | 消息通知服务 |

### 基础设施技术栈

| 组件 | 技术选型 | 说明 |
|------|----------|------|
| 容器化 | Docker + Docker Compose | 容器化部署 |
| 编排 | Kubernetes + Helm | 容器编排与管理 |
| 数据库 | PostgreSQL + MongoDB + Redis | 关系型+文档型+缓存 |
| 消息队列 | Redis + RabbitMQ | 异步消息处理 |
| 监控 | Prometheus + Grafana + Jaeger | 监控、指标、链路追踪 |
| 配置管理 | Consul + Vault | 配置中心与密钥管理 |

### 外部服务集成

| 服务 | 提供商 | 功能 |
|------|--------|------|
| ASR (语音识别) | 火山引擎 | 实时语音转文字 |
| TTS (语音合成) | 火山引擎 | 文字转语音 |
| LLM (大语言模型) | 字节豆包 | AI对话生成 |
| ISE (语音评测) | 科大讯飞 | 发音评分与分析 |

## 核心组件

### 1. API Gateway (网关层)

**职责**：
- 统一入口，路由转发

- 协议转换 (WebSocket ↔ gRPC)

- 认证授权、限流熔断

- 日志记录、监控指标

  

**技术实现**：

- 框架：Go + Gin + Gorilla WebSocket
- 协议：HTTP/WebSocket → gRPC
- 特性：连接池、健康检查、负载均衡

**关键代码路径**：
- `gateway/internal/handler/websocket.go` - WebSocket处理
- `gateway/internal/manager/websocket.go` - 连接管理
- `gateway/pkg/proto/` - gRPC客户端

### 2. Speech Service (语音服务)

**职责**：
- 音频格式转换与处理
- ASR语音识别集成
- TTS语音合成集成  
- ISE语音评测集成

**技术实现**：
- 框架：Go + gRPC + WebSocket
- 音频：PCM/WAV/WebM格式支持
- 流式处理：实时音频流处理
- 并发处理：ASR+ISE并行评测

**关键代码路径**：
- `services/speech-service/internal/handler/speech.go` - 语音处理核心
- `services/speech-service/internal/service/` - 各类语音服务
- `services/speech-service/pkg/audio/` - 音频处理工具

**数据流**：
```
Audio Input → Format Conversion → [ASR + ISE] → LLM → TTS → Audio Output
```

### 3. Conversation Service (对话服务)

**职责**：
- LLM大模型集成
- 对话上下文管理
- 个性化回复生成
- 学习内容推荐

**技术实现**：
- 框架：Python + FastAPI + gRPC
- LLM：字节豆包API集成
- 上下文：Session级别状态管理
- 个性化：用户画像与偏好学习

### 4. User Service (用户服务)

**职责**：
- 用户注册、登录、授权
- 用户画像与偏好管理
- 学习进度跟踪
- 个人数据管理

**技术实现**：
- 框架：Python + FastAPI + gRPC
- 数据库：PostgreSQL + Redis
- 认证：JWT + OAuth2
- 缓存：Redis Session存储

### 5. Analytics Service (分析服务)

**职责**：
- 学习数据统计分析
- 用户行为分析
- 性能指标监控
- 智能报表生成

**技术实现**：
- 框架：Python + Pandas + scikit-learn
- 数据处理：批处理+流处理
- 机器学习：用户行为预测
- 可视化：图表生成与展示

## 数据流设计

### 语音对话完整数据流

```
1. 用户语音输入 (Frontend)
   ↓ WebSocket Binary
2. 网关接收并转发 (Gateway)
   ↓ gRPC Stream
3. 音频处理与识别 (Speech Service)
   ├─ ASR识别 → 文本结果
   ├─ ISE评测 → 评分结果
   └─ 音频格式转换
   ↓ gRPC
4. 对话生成 (Conversation Service)
   ├─ 上下文管理
   ├─ LLM生成回复
   └─ 个性化调整
   ↓ gRPC
5. 语音合成 (Speech Service)
   ├─ TTS文本转语音
   └─ 音频编码
   ↓ gRPC Stream
6. 网关返回结果 (Gateway)
   ├─ ASR结果 (JSON)
   ├─ ISE评分 (JSON) 
   ├─ LLM回复 (JSON)
   └─ TTS音频 (Binary)
   ↓ WebSocket
7. 前端展示 (Frontend)
   ├─ 文本显示
   ├─ 评分卡片
   └─ 音频播放
```

### 数据存储设计

**PostgreSQL** (关系型数据)：
- 用户账户信息
- 学习记录与进度
- 系统配置数据

**MongoDB** (文档型数据)：
- 对话历史记录
- 用户行为日志
- 音频元数据

**Redis** (缓存数据)：
- 会话状态缓存
- 热点数据缓存
- 消息队列

## 用户交互时序图

让我为您创建一个完整的用户交互时序图：

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端页面
    participant G as Gateway
    participant S as Speech Service
    participant C as Conversation Service
    participant A as ASR API
    participant L as LLM API
    participant T as TTS API
    participant I as ISE API

    Note over U,I: 完整的语音对话流程

    U->>F: 1. 访问页面，点击连接
    F->>G: 2. WebSocket连接请求
    G-->>F: 3. 连接建立成功
    F-->>U: 4. 显示"已连接"状态

    U->>F: 5. 点击"开始录音"
    F->>F: 6. 获取麦克风权限
    F->>G: 7. 发送控制消息 (start_recording)
    G->>S: 8. gRPC开始录音指令

    Note over U,F: 用户开始说话

    U->>F: 9. 语音输入 (持续)
    F->>G: 10. WebSocket发送音频数据 (Binary)
    G->>S: 11. gRPC流式转发音频

    U->>F: 12. 点击"停止录音"
    F->>G: 13. 发送控制消息 (stop_recording)
    G->>S: 14. gRPC停止录音指令

    Note over S,I: 并行处理：ASR识别 + ISE评测

    par ASR语音识别
        S->>A: 15a. 调用ASR API
        A-->>S: 16a. 返回识别文本
        S->>G: 17a. 发送ASR结果
        G->>F: 18a. WebSocket转发ASR结果
        F-->>U: 19a. 显示识别文本
    and ISE语音评测  
        S->>I: 15b. 调用ISE API (音频+参考文本)
        I-->>S: 16b. 返回评分结果
        S->>G: 17b. 发送ISE结果
        G->>F: 18b. WebSocket转发ISE结果
        F-->>U: 19b. 显示评分卡片
    end

    Note over S,L: LLM对话生成

    S->>C: 20. 转发ASR文本到对话服务
    C->>L: 21. 调用LLM API生成回复
    L-->>C: 22. 返回AI回复文本
    C-->>S: 23. 返回处理后的回复
    S->>G: 24. 发送LLM结果
    G->>F: 25. WebSocket转发LLM结果
    F-->>U: 26. 显示AI回复文本

    Note over S,T: 语音合成

    S->>T: 27. 调用TTS API
    T-->>S: 28. 返回合成音频
    S->>G: 29. 发送TTS音频数据
    G->>F: 30. WebSocket转发音频 (Binary)
    F->>F: 31. 创建Audio元素并播放
    F-->>U: 32. 播放AI语音回复

    Note over F,U: 完整的交互循环结束，用户可以继续对话
```

### 时序图说明

这个时序图展示了一次完整的用户语音交互流程，包含以下关键阶段：

1. **连接建立** (步骤1-4): 用户连接到系统
2. **录音开始** (步骤5-8): 用户启动录音功能
3. **语音输入** (步骤9-14): 用户说话并停止录音
4. **并行处理** (步骤15-19): ASR识别和ISE评测同时进行
5. **对话生成** (步骤20-26): LLM生成智能回复
6. **语音合成** (步骤27-32): TTS生成和播放语音回复

### 并发处理优化

系统采用并发处理策略：
- **ASR + ISE 并行**: 同时进行语音识别和发音评测
- **流式处理**: 音频数据实时传输，减少延迟
- **异步响应**: 各个结果独立返回，用户体验更流畅

## API设计

### WebSocket API (前端 ↔ Gateway)

**连接格式**：
```
ws://localhost:8080/ws?session_id={session_id}
```

**消息格式**：

1. **控制消息** (JSON)
```json
{
  "type": "control",
  "data": {
    "action": "start_recording" | "stop_recording"
  },
  "session": "session_id"
}
```

2. **音频消息** (Binary)
```
WebM/PCM 格式的音频数据
```

3. **响应消息** (JSON)
```json
// ASR结果
{
  "type": "text",
  "data": {
    "type": "asr_result",
    "text": "Hello world",
    "confidence": 0.95,
    "is_final": true
  }
}

// ISE评分结果
{
  "type": "text",
  "data": {
    "type": "ise_result",
    "overall_score": 85.5,
    "accuracy_score": 80.0,
    "fluency_score": 88.0,
    "completeness_score": 90.0,
    "word_scores": [...],
    "reference_text": "Hello world"
  }
}
```

### gRPC API (内部服务通信)

**Speech Service API**：
```protobuf
service SpeechService {
  rpc ProcessVoiceConversation(stream VoiceRequest) returns (stream VoiceResponse);
}

message VoiceRequest {
  string session_id = 1;
  oneof request_type {
    AudioData audio_data = 2;
    ControlMessage control = 3;
  }
}

message VoiceResponse {
  string session_id = 1;
  oneof response_type {
    ASRResult asr_result = 2;
    LLMResult llm_result = 3;
    TTSResult tts_result = 4;
    ISEResult ise_result = 5;
    ErrorResult error = 6;
  }
}
```

### REST API (外部集成)

**用户服务API**：
```
POST /api/v1/users/register     # 用户注册
POST /api/v1/users/login        # 用户登录
GET  /api/v1/users/profile      # 获取用户资料
PUT  /api/v1/users/profile      # 更新用户资料
GET  /api/v1/users/progress     # 获取学习进度
```

**分析服务API**：
```
GET  /api/v1/analytics/stats    # 获取学习统计
GET  /api/v1/analytics/reports  # 获取学习报告
POST /api/v1/analytics/events   # 上报学习事件
```

## 部署架构

### 开发环境部署

**Docker Compose部署**：
```yaml
version: '3.8'
services:
  gateway:
    build: ./gateway
    ports: ["8080:8080"]
    depends_on: [speech-service]
    
  speech-service:
    build: ./services/speech-service
    ports: ["9090:9090"]
    environment:
      - ISE_APP_ID=${ISE_APP_ID}
      - ISE_API_KEY=${ISE_API_KEY}
    depends_on: [postgres, redis]
    
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: aiTutor
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes: ["postgres_data:/var/lib/postgresql/data"]
    
  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
```

### 生产环境部署

**Kubernetes部署**：
```yaml
# Gateway Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gateway
  template:
    metadata:
      labels:
        app: gateway
    spec:
      containers:
      - name: gateway
        image: ai-tutor/gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: SPEECH_SERVICE_URL
          value: "speech-service:9090"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

**负载均衡配置**：
```yaml
apiVersion: v1
kind: Service
metadata:
  name: gateway-service
spec:
  selector:
    app: gateway
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

### 服务发现与配置

**Consul配置**：
```json
{
  "service": {
    "name": "speech-service",
    "port": 9090,
    "tags": ["ai", "speech", "grpc"],
    "check": {
      "grpc": "localhost:9090",
      "interval": "10s"
    }
  }
}
```

**环境变量配置**：
```bash
# Speech Service
SPEECH_SERVICE_PORT=9090
ASR_API_URL=https://openspeech.bytedance.com
TTS_API_URL=https://openspeech.bytedance.com
ISE_API_URL=wss://ise-api.xfyun.cn/v2/open-ise

# Database
POSTGRES_URL=postgres://user:pass@localhost:5432/aiTutor
REDIS_URL=redis://localhost:6379/0

# External APIs
VOLCANO_ASR_APP_ID=your_app_id
VOLCANO_TTS_APP_ID=your_app_id
DOUBAO_API_KEY=your_api_key
ISE_APP_ID=your_app_id
```

## 安全设计

### 认证与授权

**JWT认证流程**：
```
1. 用户登录 → 验证凭据
2. 生成JWT Token (包含用户ID、权限、过期时间)
3. 前端存储Token，每次请求携带
4. Gateway验证Token有效性
5. 转发请求到后端服务
```

**API权限控制**：
- 公开接口：健康检查、用户注册
- 认证接口：用户登录、会话管理
- 授权接口：个人数据、学习记录

### 数据安全

**敏感数据加密**：
- 用户密码：bcrypt哈希加密
- API密钥：Vault集中管理
- 音频数据：传输加密(TLS)，存储加密

**网络安全**：
- HTTPS/WSS强制加密传输
- API限流防止攻击
- CORS跨域策略控制
- 输入验证与SQL注入防护

### 隐私保护

**音频数据处理**：
- 不持久化存储用户音频
- 处理完成后立即删除
- 外部API调用使用临时token

**用户数据保护**：
- 个人信息最小化收集
- 数据匿名化处理
- 用户数据删除权限

## 监控与日志

### 应用监控

**Prometheus指标收集**：
```yaml
# 自定义指标
- name: voice_requests_total
  help: Total number of voice requests
  type: counter
  labels: [service, method]

- name: voice_processing_duration_seconds
  help: Time spent processing voice requests
  type: histogram
  buckets: [0.1, 0.5, 1.0, 2.0, 5.0]
```

**Grafana监控面板**：
- 系统资源监控 (CPU、内存、磁盘)
- 业务指标监控 (请求量、响应时间、错误率)
- 用户行为监控 (活跃用户、学习时长)

### 日志管理

**结构化日志格式**：
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "service": "speech-service",
  "session_id": "session_123",
  "action": "asr_processing",
  "duration": 1.23,
  "status": "success"
}
```

**日志聚合**：
- ELK Stack (Elasticsearch + Logstash + Kibana)
- 集中式日志收集与分析
- 实时告警与异常检测

### 链路追踪

**Jaeger分布式追踪**：
```go
// 在Gateway中添加追踪
span := tracer.StartSpan("websocket_handler")
defer span.Finish()

// 传递追踪上下文到下游服务
ctx = opentracing.ContextWithSpan(ctx, span)
response, err := speechClient.ProcessVoice(ctx, request)
```

**追踪示例**：
```
Trace: user_voice_interaction
├─ gateway.websocket_handler (5.2s)
│  ├─ speech.process_audio (4.8s)
│  │  ├─ asr.recognize (1.2s)
│  │  ├─ ise.evaluate (1.1s)
│  │  └─ tts.synthesize (0.8s)
│  └─ conversation.generate_reply (2.1s)
```

### 告警策略

**关键指标告警**：
- 服务可用性 < 99.9%
- 响应时间 > 5秒
- 错误率 > 5%
- 内存使用率 > 80%

**业务指标告警**：
- 语音识别失败率 > 10%
- 外部API调用失败
- 用户会话异常中断

---

## 🚀 总结

AI Tutor Monorepo采用现代化的微服务架构，通过合理的技术选型和架构设计，实现了高性能、高可用、可扩展的AI英语教学平台。

### 核心优势

1. **技术先进性**: 采用云原生架构，支持容器化部署和弹性扩缩容
2. **用户体验**: 实时语音交互，智能评分反馈，现代化UI设计
3. **系统可靠性**: 微服务隔离，多重容错机制，全链路监控
4. **开发效率**: 标准化API设计，共享组件复用，自动化CI/CD

### 未来发展方向

- **AI能力增强**: 引入更先进的语言模型和语音技术
- **个性化学习**: 基于用户行为的智能推荐算法
- **多语言支持**: 扩展到更多语言的教学场景
- **社交功能**: 添加学习社区和协作功能

---

**文档版本**: v1.0  
**最后更新**: 2024年1月  
**维护团队**: CurvTech AI Team 