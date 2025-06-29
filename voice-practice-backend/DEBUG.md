# 🐛 ASR调试指南

当前版本已添加详细的调试功能，帮助诊断ASR返回空字符串的问题。

## 🎵 音频文件保存

每次处理音频时，系统会自动保存两个文件到 `debug/` 目录：

1. **原始WebM文件**: `debug/webm_[sessionID]_[timestamp].webm`
   - 前端发送的原始音频数据
   
2. **转换后的WAV文件**: `debug/audio_[sessionID]_[timestamp].wav`
   - 经过格式转换后发送给ASR的音频

## 📋 调试日志说明

### 音频处理阶段
```
🎵 Saved debug WebM file: debug/webm_session_xxx_12345.webm (1024 bytes)
🎵 Saved debug audio file: debug/audio_session_xxx_12345.wav (2048 bytes)
```

### ASR连接阶段
```
🔊 ASR Processing audio data: 2048 bytes
✅ Valid WAV file detected (RIFF+WAVE headers)
🌐 Connecting to ASR service: wss://openspeech.bytedance.com/...
🔑 ASR Headers: Access-Key=LVADY01X..., App-Key=3819780186
✅ Connected to ASR service successfully
```

### ASR响应阶段
```
📝 ASR Response for chunk 0: {...}
📊 ASR JSON payload keys: [result, confidence, is_final]
🎯 ASR Result: 'hello world' (confidence: 0.85)
```

## 🔍 排查步骤

### 1. 检查音频文件
```bash
# 查看生成的音频文件
ls -la debug/

# 试听WAV文件 (macOS)
afplay debug/audio_session_xxx_12345.wav

# 试听WAV文件 (Linux)
aplay debug/audio_session_xxx_12345.wav

# 查看WAV文件信息
file debug/audio_session_xxx_12345.wav
```

### 2. 检查日志
```bash
# 启动服务器并查看详细日志
./voice-practice-server
```

关键检查点：
- ✅ 是否检测到有效的WAV文件头？
- 🌐 是否成功连接到ASR服务？
- 📝 是否收到ASR响应？
- 📊 ASR响应的JSON结构是什么？

### 3. 常见问题诊断

#### 问题1: WAV文件无效
```
⚠️ Audio data doesn't appear to be a valid WAV file
```
**原因**: WebM到WAV转换失败  
**解决**: 检查音频转换逻辑

#### 问题2: ASR连接失败
```
❌ Failed to connect to ASR service: ...
```
**原因**: API密钥错误或网络问题  
**解决**: 检查.env配置和网络连接

#### 问题3: ASR无响应
```
📭 No payload in ASR response for chunk 0
```
**原因**: ASR服务没有返回识别结果  
**解决**: 检查音频质量和格式

#### 问题4: JSON解析失败
```
⚠️ Failed to parse ASR result from payload
```
**原因**: ASR响应格式与预期不符  
**解决**: 查看具体的JSON结构调整解析逻辑

## 🎯 测试建议

1. **录制清晰的英语语音** (2-3秒)
2. **查看生成的WAV文件** 是否可以正常播放
3. **观察服务器日志** 找到具体的错误点
4. **比对ASR demo** 确认请求格式是否正确

## 🛠️ 临时解决方案

如果音频转换有问题，可以尝试：

1. **使用命令行工具转换测试**:
```bash
# 使用ffmpeg转换WebM到WAV
ffmpeg -i debug/webm_xxx.webm -ar 16000 -ac 1 -f wav test.wav
```

2. **直接测试ASR API**:
```bash
# 使用已知好的WAV文件测试ASR
python asr_websocket_demo.py
```

## 📞 进一步调试

如果仍有问题，可以：
1. 分享生成的WAV文件
2. 提供完整的服务器日志
3. 确认API密钥是否有效

---

**注意**: 调试功能会在debug目录下保存音频文件，请注意磁盘空间使用。