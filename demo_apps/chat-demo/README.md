# 🤖 AI Tutor Chat Demo

A modern web-based chat interface for testing and demonstrating the AI Tutor conversation service. This demo provides a user-friendly way to interact with the conversation service and test its capabilities.

## 🌟 Features

- **Modern Chat Interface**: Beautiful, responsive web UI
- **Real-time Conversations**: Connect to the conversation service API
- **Multiple AI Models**: Support for GPT-4, GPT-3.5, Local, and Doubao models
- **Authentication**: JWT token-based authentication with helper tools
- **Conversation History**: Maintains context for natural conversations
- **Error Handling**: Comprehensive error handling and user feedback
- **Mobile Friendly**: Responsive design that works on mobile devices

## 🚀 Quick Start

### 1. Start the Conversation Service

First, make sure the conversation service is running:

```bash
cd ../../services/conversation-service
python main.py
```

The service should start on `http://localhost:8000` by default.

### 2. Open the Chat Demo

Simply open `index.html` in your web browser:

```bash
# Option 1: Direct file opening
open index.html  # macOS
xdg-open index.html  # Linux
start index.html  # Windows

# Option 2: Using a simple HTTP server (recommended)
python -m http.server 3000
# Then visit http://localhost:3000
```

### 3. Connect and Chat

1. **Default Connection**: The demo is pre-configured to connect to `http://localhost:8000` with the development token `dev-token`
2. **Click Connect**: Press the "Connect" button to establish connection
3. **Start Chatting**: Once connected, type your message and start conversing!

## 🔐 Authentication

The demo supports multiple authentication methods:

### Development Mode (Easiest)

Use the pre-configured development token:

- **Token**: `dev-token`
- **Works when**: Conversation service is running in debug mode
- **User**: Automatically maps to `dev-user-123`

### Custom JWT Tokens

Generate your own JWT tokens using the auth helper:

```bash
# Install PyJWT if needed
pip install PyJWT

# Create a token for a specific user
python auth_helper.py create user123 john john@example.com

# Create demo users
python auth_helper.py demo

# Validate an existing token
python auth_helper.py validate <your-token>
```

## 📋 Usage Examples

### Basic Conversation

```
User: Hello! I want to practice English conversation.
AI: Hello! I'm excited to help you practice English. What topic would you like to discuss today? We could talk about daily activities, hobbies, travel, or any subject that interests you.

User: Let's talk about travel. I went to Paris last month.
AI: How wonderful! Paris is such a beautiful city. What was your favorite part of the trip? Did you visit any famous landmarks like the Eiffel Tower or the Louvre Museum?
```

### Grammar Practice

```
User: I have went to the store yesterday.
AI: Great sentence! Just a small correction: it should be "I went to the store yesterday" or "I have been to the store today." We use simple past tense "went" with time expressions like "yesterday." What did you buy at the store?
```

### Vocabulary Building

```
User: What's another word for "happy"?
AI: There are many wonderful synonyms for "happy"! Here are some alternatives:
- Joyful (feeling great joy)
- Delighted (very pleased)
- Cheerful (in good spirits)
- Elated (extremely happy)
- Content (satisfied and peaceful)

Which one would you like to practice using in a sentence?
```

## ⚙️ Configuration

### Server Settings

You can change the server URL and authentication token in the web interface:

- **Server URL**: Default is `http://localhost:8000`
- **Auth Token**: Default is `dev-token`
- **Model Selection**: Choose from available AI models

### Model Options

The demo supports these models:

- `gpt-4-tutor`: Most advanced model for complex conversations
- `gpt-3.5-turbo-tutor`: Fast and efficient for general practice
- `local-tutor`: Local model for privacy-focused usage
- `doubao-seed-1-6-250615`: Doubao AI model (default)

## 🛠️ Development

### File Structure

```
demo_apps/chat-demo/
├── index.html          # Main chat interface
├── auth_helper.py      # JWT token generation tool
├── README.md          # This documentation
└── server.py          # Optional local server (if needed)
```

### Customization

#### Styling

All styles are contained within `index.html` in the `<style>` section. Key customization points:

- **Colors**: Modify the gradient backgrounds in `.chat-header` and `.send-button`
- **Layout**: Adjust `.chat-container` dimensions and `.message` styling
- **Typography**: Change font families and sizes in the universal selector

#### Functionality

JavaScript functionality is in the `<script>` section:

- **ChatDemo class**: Main application logic
- **API calls**: Modify `callChatAPI()` for different endpoints
- **Message handling**: Customize `addMessage()` for different message types

### Adding Features

#### Streaming Support

The conversation service supports streaming responses. To enable:

1. Modify the API call to include `stream: true`
2. Handle Server-Sent Events (SSE) in the response
3. Update the UI progressively as chunks arrive

#### File Upload

Add image/file upload support for multimodal conversations:

1. Add file input to the chat interface
2. Use the `/conversation/multimodal` endpoint
3. Handle base64 encoding for images

## 🔧 Troubleshooting

### Connection Issues

**Problem**: "Failed to connect" error

- **Solution**: Ensure conversation service is running on `http://localhost:8000`
- **Check**: Service logs for startup errors
- **Verify**: No firewall blocking port 8000

**Problem**: CORS errors in browser console

- **Solution**: Conversation service has CORS enabled for all origins
- **Check**: Browser developer tools for specific CORS errors
- **Try**: Using a local HTTP server instead of file:// protocol

### Authentication Issues

**Problem**: "Invalid authentication credentials" error

- **Solution**: Check if using correct token format
- **Verify**: Token hasn't expired (24 hours default)
- **Try**: Using `dev-token` for development

**Problem**: Token validation fails

- **Solution**: Ensure conversation service is in debug mode for `dev-token`
- **Check**: JWT secret matches between auth helper and service
- **Generate**: New token using auth helper

### API Issues

**Problem**: Model not available

- **Solution**: Check `/v1/models` endpoint to see available models
- **Verify**: Selected model is supported by the service
- **Check**: Service configuration for external API keys

**Problem**: Slow responses

- **Solution**: Try different models (local-tutor for fastest)
- **Check**: Service resource usage and external API status
- **Consider**: Shorter messages for faster processing

## 🤝 Contributing

Want to improve the demo? Here are some ideas:

1. **Add streaming support** for real-time response rendering
2. **Implement conversation export** to save chat history
3. **Add voice input/output** for speaking practice
4. **Create conversation templates** for specific learning scenarios
5. **Add progress tracking** to monitor learning improvement

## 📝 License

This demo is part of the AI Tutor monorepo and follows the same license terms.

## 🙋‍♀️ Support

- **Service Issues**: Check conversation service logs and documentation
- **Demo Issues**: Review browser console for JavaScript errors
- **API Issues**: Test endpoints directly using curl or Postman
- **Authentication**: Use auth helper tool to validate tokens

---

Happy learning! 🎓✨
