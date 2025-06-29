# 🕵️ 调试步骤指南

## 🎯 当前问题
- 前端录音后生成的blob URL访问404
- debug目录没有文件生成
- 说明音频数据没有到达后端

## 🔍 调试步骤

### 1. 启动服务器并查看连接日志
```bash
./voice-practice-server
```

**期望看到的日志**:
```
🌐 WebSocket connection request for session: session_xxx
✅ WebSocket connection established for session: session_xxx
```

### 2. 打开浏览器开发者工具
- 打开 `http://localhost:8080`
- 按F12打开开发者工具
- 点击"Connect"按钮

**期望看到的前端日志**:
```
Connected to server successfully
```

### 3. 测试录音流程
点击"Start Recording"，说话3-5秒，然后点击"Stop Recording"

**期望看到的前端日志顺序**:
```
✅ Using WebM with Opus codec
🎬 MediaRecorder started
🎤 Audio chunk received: 1024 bytes
🎤 Audio chunk received: 512 bytes
... (更多audio chunks)
🛑 stopRecording called, isRecording: true
📹 MediaRecorder state before stop: recording
📹 MediaRecorder state after stop: inactive
🎬 MediaRecorder onstop triggered
📦 Total chunks collected: 15
🎵 Complete audio blob: 15360 bytes, type: audio/webm;codecs=opus
🔊 Test audio URL (paste in browser): blob:http://localhost:8080/12345...
🌐 WebSocket state: 1
📤 Sending audio blob to server...
```

**期望看到的服务器日志**:
```
📨 Received message for session session_xxx: type=2, size=15360
🎵 Processing binary message for session session_xxx (15360 bytes)
🔥 handleBinaryMessage called for session session_xxx with 15360 bytes
✅ Session found, IsRecording: true
🎵 Processing complete audio file for session session_xxx: 15360 bytes
🎬 Processing complete audio file for session session_xxx (15360 bytes)
🎵 Saved debug WebM file: debug/webm_COMPLETE_session_xxx_12345.webm (15360 bytes)
```

## 🚨 可能的问题和解决方案

### 问题1: 没有看到WebSocket连接日志
**可能原因**: 服务器没有启动或端口被占用
**解决方案**: 
```bash
# 检查端口
lsof -i :8080
# 停止占用端口的进程
sudo kill -9 <PID>
```

### 问题2: 前端显示连接失败
**可能原因**: WebSocket连接失败
**解决方案**: 检查浏览器控制台错误信息

### 问题3: MediaRecorder没有触发onstop
**可能原因**: 浏览器兼容性问题
**解决方案**: 
1. 确认麦克风权限已授权
2. 尝试不同浏览器 (Chrome/Firefox)
3. 检查是否有其他应用占用麦克风

### 问题4: 收集到的audio chunks为0
**可能原因**: 麦克风没有真正录音
**解决方案**: 
1. 检查系统麦克风设置
2. 确认页面有HTTPS或localhost访问
3. 尝试其他麦克风设备

### 问题5: WebSocket状态不是1 (OPEN)
**可能原因**: WebSocket连接断开
**解决方案**: 
1. 检查网络连接
2. 重新刷新页面建立连接

### 问题6: blob URL访问404
**期望**: blob URL应该是 `blob:http://localhost:8080/xxx-xxx-xxx` 格式
**如果出现404**: 说明blob创建失败，检查音频数据是否有效

## 🎯 关键检查点

1. **WebSocket连接** ✅ 
   - 看到服务器连接日志
   - 前端显示"Connected"状态

2. **音频录制** ✅
   - 看到多个"Audio chunk received"
   - MediaRecorder状态变化正常

3. **音频发送** ✅
   - 看到"Sending audio blob to server"
   - 服务器收到binary message

4. **文件保存** ✅
   - debug目录生成文件
   - 文件大小>0

按照这个顺序逐步检查，找到具体哪一步出现问题！