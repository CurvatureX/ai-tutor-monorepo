# AI English Speaking Practice Backend

A real-time voice interaction backend for English speaking practice, built with Go and WebSocket technology.

## Features

- **Real-time Audio Processing**: WebSocket-based audio streaming with WebM to PCM/WAV conversion
- **Speech Recognition**: Integration with Volcano Engine ASR service for speech-to-text
- **AI Conversation**: LLM-powered English conversation practice with contextual responses
- **Speech Synthesis**: Text-to-speech conversion for AI responses
- **Web Interface**: Simple, responsive frontend for voice interaction

## Architecture

```
Frontend (WebRTC) → WebSocket → Audio Processing → ASR → LLM → TTS → Audio Response
```

### Tech Stack

- **Backend**: Go with Gin framework
- **WebSocket**: gorilla/websocket for real-time communication
- **Audio Processing**: Custom audio format conversion (WebM ↔ PCM/WAV)
- **External APIs**: Volcano Engine (ASR, LLM, TTS services)
- **Frontend**: Vanilla JavaScript with WebRTC APIs

## Quick Start

### 1. Prerequisites

- Go 1.21 or later
- Volcano Engine API credentials (ASR, LLM, TTS)

### 2. Installation

```bash
cd voice-practice-backend
go mod tidy
```

### 3. Configuration

The application automatically loads configuration from a `.env` file in the project root. A template file `.env.example` is provided.

Edit the `.env` file with your Volcano Engine API keys:

```env
# Server Configuration
PORT=8080
HOST=localhost

# ASR Service Configuration (Volcano Engine)
ASR_API_KEY=your_actual_asr_api_key
ASR_API_SECRET=your_actual_asr_api_secret

# LLM Service Configuration (Volcano Engine)  
LLM_API_KEY=your_actual_llm_api_key
LLM_API_SECRET=your_actual_llm_api_secret

# TTS Service Configuration (Volcano Engine)
TTS_API_KEY=your_actual_tts_api_key
TTS_API_SECRET=your_actual_tts_api_secret

# Audio settings can usually keep defaults
AUDIO_SAMPLE_RATE=16000
AUDIO_CHANNELS=1
```

**Note**: The application will work without a `.env` file by falling back to system environment variables, but using `.env` is recommended for development.

### 4. Run the Server

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

### 5. Access the Web Interface

Open your browser and navigate to `http://localhost:8080`

## API Endpoints

- `GET /` - Web interface
- `GET /ws` - WebSocket endpoint for voice interaction
- `GET /health` - Health check
- `GET /ready` - Readiness check

## Usage

1. **Connect**: Click "Connect" to establish WebSocket connection
2. **Record**: Click "Start Recording" and speak in English
3. **AI Response**: The system will:
   - Convert speech to text (ASR)
   - Generate conversational response (LLM)
   - Convert response to speech (TTS)
   - Play the audio response

## WebSocket Protocol

### Client → Server Messages

#### Control Messages (JSON)
```json
{
  "type": "control",
  "data": {"action": "start_recording|stop_recording|end_session"},
  "session": "session_id"
}
```

#### Audio Data (Binary)
- Raw WebM audio chunks from MediaRecorder

### Server → Client Messages

#### Text Messages (JSON)
```json
{
  "type": "text|error",
  "data": "message_content_or_structured_data",
  "session": "session_id"
}
```

#### Audio Data (Binary)
- MP3 audio data for TTS responses

## Audio Processing Pipeline

1. **Input**: WebM (Opus codec) from browser MediaRecorder
2. **Conversion**: WebM → PCM → WAV for ASR compatibility
3. **Recognition**: WAV → Text via Volcano ASR
4. **Generation**: Text → LLM Response
5. **Synthesis**: Response Text → MP3 via Volcano TTS
6. **Output**: MP3 audio streamed to browser

## Development

### Project Structure

```
voice-practice-backend/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP/WebSocket handlers
│   ├── model/           # Data structures
│   └── service/         # Business logic (ASR, LLM, TTS, Audio)
├── pkg/
│   ├── audio/           # Audio conversion utilities
│   └── websocket/       # WebSocket connection management
└── static/              # Frontend files
```

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o voice-practice-server cmd/server/main.go
```

## Configuration Options

All configuration can be set via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Server port |
| `HOST` | localhost | Server host |
| `ASR_API_KEY` | - | Volcano ASR API key |
| `LLM_API_KEY` | - | Volcano LLM API key |
| `TTS_API_KEY` | - | Volcano TTS API key |
| `AUDIO_SAMPLE_RATE` | 16000 | Audio sample rate |
| `AUDIO_CHANNELS` | 1 | Audio channels (mono) |

## Troubleshooting

### Audio Issues
- Ensure browser has microphone permissions
- Check audio format compatibility (WebM/Opus support)
- Verify sample rate settings (16kHz recommended)

### Connection Issues
- Verify WebSocket URL and port
- Check CORS settings for cross-origin requests
- Ensure API credentials are correctly configured

### Performance
- Monitor WebSocket message frequency (100ms chunks recommended)
- Adjust buffer sizes for better latency/quality tradeoff
- Consider implementing audio compression for bandwidth optimization

## License

MIT License - see LICENSE file for details.