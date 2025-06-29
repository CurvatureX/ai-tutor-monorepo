# 🎯 WebM转WAV解决方案

## 🔍 问题分析
- ✅ WebM文件录制和传输正常
- ❌ WAV转换产生白噪音 
- **原因**: 当前使用的是模拟转换器，不是真正的WebM解码

## 💡 解决方案

### 方案1: 安装ffmpeg (推荐)

```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt update && sudo apt install ffmpeg

# CentOS/RHEL
sudo yum install ffmpeg
```

安装后重启服务器，应该看到：
```
✅ ffmpeg detected - using real WebM conversion
```

### 方案2: 使用Docker运行 (如果本地安装困难)

```bash
# 拉取包含ffmpeg的镜像
docker run -it --rm \
  -v $(pwd):/app \
  -w /app \
  -p 8080:8080 \
  jrottenberg/ffmpeg:alpine \
  /bin/sh

# 在容器内运行
./voice-practice-server
```

### 方案3: 临时测试 - 手动转换

如果暂时无法安装ffmpeg，可以手动验证转换：

```bash
# 1. 运行服务器，录制音频
./voice-practice-server

# 2. 手动转换保存的WebM文件
ffmpeg -i debug/webm_COMPLETE_*.webm -ar 16000 -ac 1 -f wav test_manual.wav

# 3. 播放测试
afplay test_manual.wav  # macOS
aplay test_manual.wav   # Linux
```

如果手动转换的WAV文件播放正常，说明方案有效。

### 方案4: 在线转换测试

1. 将保存的WebM文件上传到在线转换工具
2. 转换为WAV格式
3. 下载并播放验证

## 🚀 验证步骤

### 1. 检查ffmpeg状态
启动服务器后查看日志：
```
✅ ffmpeg detected - using real WebM conversion  # 正常
⚠️ ffmpeg not found - using simulated conversion  # 需要安装
```

### 2. 测试转换效果
录制音频后检查：
```bash
# WebM文件应该能播放
afplay debug/webm_COMPLETE_*.webm

# WAV文件应该也能播放 (安装ffmpeg后)
afplay debug/audio_COMPLETE_*.wav
```

### 3. ASR测试
如果WAV文件正常，ASR应该能正确识别语音内容。

## 📋 故障排除

### 问题1: ffmpeg安装失败
**解决**: 
- 尝试不同的包管理器
- 使用Docker方案
- 从源码编译

### 问题2: 权限问题
**解决**:
```bash
# 确保临时目录有写权限
chmod 755 /tmp
```

### 问题3: ffmpeg版本问题
**验证**:
```bash
ffmpeg -version
# 应该支持WebM和Opus解码
```

### 问题4: 仍然产生白噪音
**检查**:
1. 确认WebM文件本身可播放
2. 检查ffmpeg是否真正被调用
3. 查看服务器日志中的转换错误

## 🎯 预期结果

正确配置后：
1. ✅ WebM文件播放正常
2. ✅ WAV文件播放正常  
3. ✅ ASR识别准确
4. ✅ 整个语音交互流程完整

ffmpeg是处理WebM/Opus音频的标准工具，安装后应该完全解决转换问题！