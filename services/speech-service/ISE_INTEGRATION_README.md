# 语音评测功能集成 (ISE - Intelligent Speech Evaluation)

## 概述

本项目已成功集成讯飞语音评测API，为AI英语口语练习系统添加了智能发音评测功能。该功能基于讯飞开放平台的[语音评测流式版API](https://www.xfyun.cn/doc/Ise/IseAPI.html)实现。

## 功能特性

### 🎯 核心评测功能
- **综合评分**: 0-100分的总体发音质量评分
- **多维度评测**: 准确度、流利度、完整度分别评分
- **多级别分析**: 支持音素、单词、句子三个级别的详细分析

### 🔧 技术特点
- **实时流式处理**: 基于WebSocket的实时音频传输
- **HMAC-SHA256认证**: 安全的API认证机制
- **自动分类识别**: 智能识别音节、单词、句子、段落类型
- **多语言支持**: 支持中文(zh_cn)和英文(en_us)评测

## 架构设计

```
客户端音频 → AudioService → ISEService → 讯飞API
                ↓
         VoiceResponse ← ISEResult ← 评测结果
```

## 配置说明

### 环境变量配置
```bash
# ISE服务配置
ISE_APP_ID=your_app_id
ISE_API_KEY=your_api_key  
ISE_API_SECRET=your_api_secret
ISE_BASE_URL=wss://ise-api.xfyun.cn/v2/open-ise
ISE_LANGUAGE=en_us  # 或 zh_cn
```

### 配置结构
```go
type ISEConfig struct {
    AppID     string // 讯飞开放平台应用ID
    APIKey    string // API密钥
    APISecret string // API密钥
    BaseURL   string // WebSocket连接地址
    Language  string // 评测语言 "zh_cn" or "en_us"
}
```

## API使用示例

### 1. 基本评测请求
```go
request := &model.ISERequest{
    AudioData: audioBytes,           // WAV/PCM音频数据
    Text:      "Hello, how are you?", // 参考文本
    Language:  "en_us",              // 评测语言
    Category:  "read_sentence",      // 可选，会自动推断
}

response, err := iseService.EvaluateSpeech(request)
```

### 2. 评测结果结构
```go
type ISEResponse struct {
    OverallScore      float64         // 总分 0-100
    AccuracyScore     float64         // 准确度分数
    FluencyScore      float64         // 流利度分数  
    CompletenessScore float64         // 完整度分数
    WordScores        []WordScore     // 单词级评分
    PhoneScores       []PhoneScore    // 音素级评分
    SentenceScores    []SentenceScore // 句子级评分
    IsFinal           bool            // 是否为最终结果
}
```

### 3. gRPC流式调用
```go
// 客户端发送音频数据
request := &speechv1.VoiceRequest{
    SessionId: "session_123",
    RequestType: &speechv1.VoiceRequest_AudioData{
        AudioData: &speechv1.AudioData{
            Data: audioBytes,
            Format: &speechv1.AudioFormat{
                Codec:      "wav",
                SampleRate: 16000,
                Channels:   1,
                BitDepth:   16,
            },
        },
    },
}

// 服务器返回ISE评测结果
response := &speechv1.VoiceResponse{
    ResponseType: &speechv1.VoiceResponse_IseResult{
        IseResult: &speechv1.ISEResult{
            OverallScore: 85.5,
            AccuracyScore: 88.0,
            FluencyScore: 82.0,
            // ... 详细评分数据
        },
    },
}
```

## 评测类别说明

系统会根据输入文本自动判断评测类别：

| 类别 | 适用场景 | 示例 |
|------|----------|------|
| `read_syllable` | 音节评测 | "ba", "to" |
| `read_word` | 单词评测 | "hello", "pronunciation" |
| `read_sentence` | 句子评测 | "How are you today?" |
| `read_chapter` | 段落评测 | 长段落文本 |

## 详细评分说明

### 单词级评分 (WordScore)
```go
type WordScore struct {
    Word       string  // 单词内容
    Score      float64 // 单词发音分数 0-100
    StartTime  int64   // 开始时间(毫秒)
    EndTime    int64   // 结束时间(毫秒) 
    IsCorrect  bool    // 是否发音正确
    Confidence float64 // 置信度
}
```

### 音素级评分 (PhoneScore)
```go
type PhoneScore struct {
    Phone     string  // 音素符号 如 /ə/, /t/, /h/
    Score     float64 // 音素发音分数
    StartTime int64   // 开始时间(毫秒)
    EndTime   int64   // 结束时间(毫秒)
    IsCorrect bool    // 是否发音正确(>50分)
}
```

## 集成流程

### 1. 服务初始化
```go
// 加载配置
cfg := config.Load()

// 创建ISE服务
iseService := service.NewISEService(&cfg.ISE, logger)

// 集成到语音处理器
speechHandler := handler.NewSpeechHandler(
    audioService, asrService, llmService, 
    ttsService, iseService, logger)
```

### 2. 音频处理流程
1. **音频接收**: 客户端通过WebSocket发送音频数据
2. **格式转换**: AudioService优化音频格式(16kHz, 16bit, mono)
3. **并行处理**: ASR识别 + ISE评测同时进行
4. **结果返回**: 通过gRPC流返回评测结果

## 最佳实践

### 🎵 音频质量要求
- **采样率**: 16kHz (必需)
- **位深度**: 16bit
- **声道**: 单声道(mono)
- **格式**: WAV, PCM (推荐)
- **时长**: 不超过5分钟

### 📝 参考文本建议
- 使用标准拼写和标点
- 避免特殊符号和数字
- 中文使用简体中文
- 英文建议使用常见词汇

### ⚡ 性能优化
- 音频分块大小: 1280字节 (~40ms)
- 并发评测: ASR + ISE 同时处理
- 连接复用: 单次WebSocket连接处理完整评测

## 错误处理

常见错误及解决方案:

| 错误类型 | 原因 | 解决方案 |
|----------|------|----------|
| 认证失败 | API密钥错误 | 检查AppID/APIKey/APISecret |
| 音频格式错误 | 格式不支持 | 转换为16kHz WAV格式 |
| 文本为空 | 缺少参考文本 | 提供标准参考文本 |
| 网络超时 | 连接不稳定 | 检查网络连接和防火墙 |

## 监控指标

系统提供以下监控指标:
- ISE评测请求数量
- 评测成功率
- 平均评测时间
- 错误类型分布
- 音频质量分布

## 示例日志输出

```
INFO[2024-06-29T15:59:30Z] 🎯 ISE Processing: 32000 bytes audio, text: 'Hello, how are you?', language: en_us
INFO[2024-06-29T15:59:31Z] 🔐 Connecting to ISE with auth: wss://ise-api.xfyun.cn/v2/open-ise?authorization=...
INFO[2024-06-29T15:59:31Z] ✅ Connected to ISE service successfully
DEBUG[2024-06-29T15:59:32Z] 📊 ISE partial result for chunk 0: score 85.50
INFO[2024-06-29T15:59:32Z] ✅ ISE Evaluation complete: overall score 85.50
INFO[2024-06-29T15:59:32Z] Generated ISE evaluation for session session_123: overall score 85.50
```

## 后续扩展

未来可以考虑的功能扩展:
- **语调评测**: 分析语音的节奏和语调
- **情感分析**: 识别语音中的情感色彩
- **个性化建议**: 基于评测结果提供学习建议
- **进度跟踪**: 记录学习者的发音进步轨迹
- **对比分析**: 与标准发音进行可视化对比

---

## 技术支持

如有问题，请参考:
- [讯飞语音评测API文档](https://www.xfyun.cn/doc/Ise/IseAPI.html)
- 项目Issue: [GitHub Issues](/)
- 技术交流群: 待建立

**版本**: v1.0.0  
**更新时间**: 2024-06-29 