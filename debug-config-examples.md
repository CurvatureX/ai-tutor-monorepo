# 🐛 Debug Configuration Examples

This file contains configuration examples for connecting your IDE to the Delve debugger when running services in debug mode.

## 🚀 Starting Services in Debug Mode

```bash
./start-local-debug.sh
# Choose option 2 for Delve debug mode
```

This will start:
- **Speech Service**: Delve debugger on `localhost:2345`
- **Gateway**: Delve debugger on `localhost:2346`

## 🔧 IDE Configuration Examples

### VS Code (.vscode/launch.json)

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Speech Service (Remote)",
            "type": "go",
            "request": "attach",
            "mode": "remote",
            "remotePath": "${workspaceFolder}/services/speech-service",
            "port": 2345,
            "host": "127.0.0.1",
            "showLog": true,
            "trace": "verbose"
        },
        {
            "name": "Debug Gateway (Remote)",
            "type": "go",
            "request": "attach",
            "mode": "remote",
            "remotePath": "${workspaceFolder}/gateway",
            "port": 2346,
            "host": "127.0.0.1",
            "showLog": true,
            "trace": "verbose"
        }
    ]
}
```

### GoLand/IntelliJ IDEA

1. **Run/Debug Configurations** → **Go Remote**
2. **Speech Service Configuration:**
   - Name: `Debug Speech Service`
   - Host: `localhost`
   - Port: `2345`
   - On disconnect: `Stop remote Delve process`

3. **Gateway Configuration:**
   - Name: `Debug Gateway`
   - Host: `localhost`
   - Port: `2346`
   - On disconnect: `Stop remote Delve process`

### Vim/Neovim with vim-go

```vim
" Add to your .vimrc or init.vim
let g:go_debug_address = 'localhost:2345'

" Connect to debugger
:GoDebugConnect localhost:2345
```

### Sublime Text with GoSublime

```json
{
    "name": "Debug Speech Service",
    "cmd": ["dlv", "connect", "localhost:2345"],
    "working_dir": "${project_path}/services/speech-service"
}
```

## 🎯 Usage Instructions

### 1. Start Debug Mode
```bash
./start-local-debug.sh
# Choose option 2: Delve Debug
```

### 2. Set Breakpoints
In your IDE, set breakpoints in the Go source files:
- **Speech Service**: `services/speech-service/internal/handler/speech.go`
- **Gateway**: `gateway/internal/handler/websocket.go`

### 3. Connect Debugger
- Use your IDE's "Attach to Remote Process" or "Connect to Delve" feature
- Connect to `localhost:2345` for Speech Service
- Connect to `localhost:2346` for Gateway

### 4. Test with Breakpoints
- Open browser: http://localhost:8080
- Start recording voice
- Your breakpoints should trigger!

## 🔍 Common Debugging Scenarios

### Debug ISE WebSocket Issues
Set breakpoints in:
```
services/speech-service/internal/service/ise.go:310 (sendMultipleAudioChunks)
services/speech-service/internal/service/ise.go:449 (sendAudioChunk)
```

### Debug Audio Processing
Set breakpoints in:
```
services/speech-service/internal/handler/speech.go:206 (processCompleteAudio)
services/speech-service/pkg/audio/converter.go (convertWebMToPCM)
```

### Debug WebSocket Connection
Set breakpoints in:
```
gateway/internal/handler/websocket.go (handleWebSocketConnection)
gateway/internal/manager/websocket.go (broadcast methods)
```

## 🛠️ Troubleshooting

### Issue: "Failed to connect to debugger"
**Solution:**
1. Make sure services are running in delve mode
2. Check ports 2345/2346 are not blocked
3. Verify delve is installed: `dlv version`

### Issue: "Breakpoints not hit"
**Solution:**
1. Ensure source paths match in IDE configuration
2. Check that breakpoints are set in the correct service
3. Verify the code path is actually executed

### Issue: "Debugger disconnects"
**Solution:**
1. Check the terminal running debug services for errors
2. Increase timeout settings in your IDE
3. Use `--accept-multiclient` flag (already included)

## 💡 Tips

- **Multiple Connections**: Both debuggers support multiple IDE connections
- **Log Files**: Even in debug mode, logs are saved to files for reference
- **Source Maps**: Debug symbols are included for better stack traces
- **Hot Reload**: Restart services to apply code changes

## 🛑 Stopping Debug Mode

```bash
# In the terminal running debug services
Ctrl+C

# Or from another terminal
./stop-local-debug.sh
``` 