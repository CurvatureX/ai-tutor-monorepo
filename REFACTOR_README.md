# Voice Practice Backend 重构完成

## 📋 重构概述

成功将 `voice-practice-backend` 重构为两个独立的微服务：

### 🌐 Gateway服务 (`/gateway`)
- **功能**: WebSocket网关，负责协议转换和流量转发
- **端口**: 8080
- **职责**:
  - WebSocket连接管理
  - 协议转换 (WebSocket ↔ gRPC)
  - 会话状态管理
  - 静态文件服务

### 🎙️ Speech Service (`/services/speech-service`)
- **功能**: 流式语音处理gRPC服务
- **端口**: 50051
- **职责**:
  - 音频格式转换 (WebM → WAV)
  - 语音识别 (ASR)
  - 对话生成 (LLM)
  - 语音合成 (TTS)

## 🏗️ 架构设计

```
Client (WebSocket) → Gateway → Speech Service (gRPC) → External APIs
                 ↓              ↓
            协议转换         语音处理管道
         (WebSocket↔gRPC)   (ASR→LLM→TTS)
```

### 核心接口
- **gRPC接口**: `ProcessVoiceConversation` (双向流式)
- **WebSocket接口**: `/ws` (与原系统保持兼容)

## 🚀 快速启动

### 1. 使用Docker启动
```bash
docker-compose up
```

### 2. 手动启动
```bash
# 启动Speech Service
cd services/speech-service
./speech-service

# 启动Gateway
cd gateway
./gateway
```

### 3. 测试连接
- WebSocket测试页面: http://localhost:8080
- 健康检查: http://localhost:8080/health
- 就绪检查: http://localhost:8080/ready

## 📁 目录结构

```
voice-practice-backend/         # 原始代码(已重构)
├── gateway/                    # WebSocket网关服务
│   ├── internal/
│   │   ├── config/            # 配置管理
│   │   ├── handler/           # WebSocket处理器
│   │   ├── manager/           # 连接管理器
│   │   └── model/             # 数据模型
│   ├── pkg/proto/speech/      # 生成的gRPC代码
│   ├── static/                # 静态文件
│   └── main.go
│
├── services/speech-service/    # 语音处理服务
│   ├── internal/
│   │   ├── config/            # 配置管理
│   │   ├── handler/           # gRPC处理器
│   │   ├── model/             # 数据模型
│   │   └── service/           # 业务逻辑
│   ├── pkg/proto/speech/      # 生成的gRPC代码
│   ├── cmd/                   # 主程序
│   └── go.mod
│
└── shared/proto/speech/        # Protocol Buffer定义
    └── speech.proto
```

## 🔧 配置说明

### Gateway环境变量
```bash
GATEWAY_HOST=0.0.0.0           # 服务监听地址
GATEWAY_PORT=8080              # 服务端口
SPEECH_SERVICE_HOST=localhost  # Speech Service地址
SPEECH_SERVICE_PORT=50051      # Speech Service端口
LOG_LEVEL=info                 # 日志级别
```

### Speech Service环境变量
```bash
SERVER_HOST=0.0.0.0           # 服务监听地址
SERVER_PORT=50051             # 服务端口

# ASR配置
ASR_ACCESS_KEY=your_key       # 字节跳动ASR访问密钥
ASR_APP_KEY=your_app_key      # 字节跳动ASR应用密钥
ASR_URL=wss://openspeech.bytedance.com/api/v2/asr

# LLM配置
LLM_API_KEY=your_api_key      # OpenAI API密钥
LLM_MODEL=gpt-3.5-turbo       # 模型名称
LLM_URL=https://api.openai.com/v1/chat/completions

# TTS配置
TTS_APP_ID=your_app_id        # 字节跳动TTS应用ID
TTS_TOKEN=your_token          # 字节跳动TTS令牌
TTS_CLUSTER=volcano_tts       # TTS集群
TTS_VOICE_TYPE=BV001_streaming # 语音类型
```

## 🛠️ 开发指南

### 重新生成Proto代码
```bash
python3 tools/proto-gen/generate.py
```

### 构建服务
```bash
# 构建Gateway
cd gateway && go build -o gateway .

# 构建Speech Service
cd services/speech-service && go build -o speech-service ./cmd
```

### 运行测试
```bash
./test-refactor.sh
```

## 📈 优势与改进

### ✅ 架构优势
1. **服务解耦**: Gateway和Speech Service各司其职
2. **协议标准化**: 使用gRPC提供类型安全的接口
3. **可扩展性**: 支持水平扩展和负载均衡
4. **容器化**: 提供完整的Docker支持
5. **向后兼容**: WebSocket接口保持不变

### 🔄 迁移说明
- 客户端无需修改，WebSocket接口完全兼容
- 所有原有功能保持不变
- 配置方式保持一致
- 调试和日志功能增强

### 🚀 扩展可能
1. 添加新的语音处理功能
2. 支持多种音频格式
3. 实现负载均衡
4. 添加监控和指标
5. 支持多语言客户端

## 📞 联系与支持

如有问题，请查看：
1. Gateway日志: `docker logs <gateway_container>`
2. Speech Service日志: `docker logs <speech_service_container>`
3. 健康检查接口验证服务状态

重构成功完成！🎉