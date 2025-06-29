# 🎯 Demo Applications

This directory contains demo applications that showcase the AI Tutor conversation service capabilities. These demos are designed to help developers understand how to integrate with the conversation service and provide examples for different use cases.

## 📱 Available Demos

### 🤖 Chat Demo (`chat-demo/`)

A modern, web-based chat interface for testing and demonstrating the AI Tutor conversation service.

**Features:**

- Modern responsive web UI
- Real-time conversations with AI tutors
- Support for multiple AI models (GPT-4, GPT-3.5, Local, Doubao)
- JWT authentication with helper tools
- Conversation history and context
- Error handling and connection management

**Quick Start:**

```bash
cd chat-demo/
python setup.py      # Automated setup
# or
python server.py     # Manual server start
```

[📖 Full Documentation](./chat-demo/README.md)

## 🚀 Getting Started

### Prerequisites

- Python 3.7 or higher
- AI Tutor conversation service running (default: `localhost:8000`)

### Quick Setup for Any Demo

1. **Start the conversation service:**

   ```bash
   cd ../services/conversation-service
   python main.py
   ```

2. **Choose and run a demo:**

   ```bash
   cd demo_apps/chat-demo/
   python setup.py
   ```

3. **Open your browser** and start chatting!

## 🔧 Development

### Adding New Demos

To add a new demo application:

1. **Create a new directory** under `demo_apps/`
2. **Include these essential files:**
   - `README.md` - Documentation
   - `requirements.txt` - Dependencies (if any)
   - Main application files
3. **Update this README** to list the new demo
4. **Follow the patterns** established by existing demos

### Demo Structure Guidelines

```
demo_apps/
├── your-demo-name/
│   ├── README.md           # Complete documentation
│   ├── requirements.txt    # Python dependencies
│   ├── setup.py           # Optional setup script
│   ├── main files...      # Core application
│   └── assets/            # Static files (if needed)
└── README.md              # This file
```

## 🤝 Contributing

We welcome contributions of new demo applications! Ideas for demos:

- **Mobile app demo** (React Native/Flutter)
- **CLI chat tool** (Python/Node.js)
- **Voice conversation demo** (Speech-to-text/Text-to-speech)
- **Educational game** (Language learning game)
- **Slack/Discord bot** (Chat platform integration)
- **REST API examples** (Various programming languages)

## 📝 License

All demo applications follow the same license as the main AI Tutor project.

## 🆘 Support

- **Demo Issues**: Check individual demo README files
- **Service Issues**: See `../services/conversation-service/README.md`
- **General Help**: Open an issue in the main repository

---

Happy coding! 🎓✨
