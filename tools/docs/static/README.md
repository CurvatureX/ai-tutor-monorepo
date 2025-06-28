# AI Tutor Monorepo

A comprehensive AI-powered tutoring platform built with microservices architecture, supporting multiple client applications and real-time learning experiences.

在线文档： https://curvaturex.github.io/ai-tutor-monorepo

## 🏗️ Architecture Overview

This monorepo contains a complete AI tutoring ecosystem with:

- **Multiple Client Applications**: Flutter mobile app and Unity 3D immersive learning environments
- **Microservices Backend**: Scalable services for user management, conversations, speech processing, analytics, and notifications
- **Shared Infrastructure**: Common protocols, configurations, and utilities
- **DevOps & Deployment**: Kubernetes, Helm charts, monitoring, and CI/CD pipelines

## 📱 Client Applications

### Flutter Mobile App

- Cross-platform mobile application for iOS and Android
- Real-time chat interface with AI tutors
- Voice interaction capabilities
- Progress tracking and analytics

### Unity 3D Application

- Immersive 3D learning environments
- Interactive educational content
- Gamified learning experiences

## 🔧 Backend Services

### Core Services

- **User Service**: Authentication, user profiles, and account management
- **Conversation Service**: AI-powered chat and tutoring logic
- **Speech Service**: Voice recognition and text-to-speech processing
- **Analytics Service**: Learning progress tracking and insights
- **Notification Service**: Real-time alerts and messaging

### Infrastructure

- **API Gateway**: Request routing and load balancing
- **Shared Protocols**: gRPC definitions and common data models
- **Monitoring**: Observability and performance tracking
- **Database**: User data, conversation history, and analytics storage

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Kubernetes cluster (for production)
- Go 1.19+ (for gateway and notification service)
- Python 3.9+ (for AI services)
- Flutter SDK (for mobile development)
- Unity 2022.3+ (for 3D application)

### Development Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/your-org/ai-tutor-monorepo.git
   cd ai-tutor-monorepo
   ```

2. **Start backend services**

   ```bash
   docker-compose up -d
   ```

3. **Run Flutter app**

   ```bash
   cd clients/flutter-app
   flutter pub get
   flutter run
   ```

4. **Open Unity project**
   ```
   # Open clients/unity-3d in Unity Editor
   ```

## Development

### Project Structure

```
├── clients/           # Client applications
│   ├── flutter-app/   # Mobile application
│   └── unity-3d/      # 3D learning environment
├── services/          # Backend microservices
├── gateway/           # API gateway
├── shared/            # Common code and protocols
├── infrastructure/    # Deployment configurations
└── tools/             # Development and build tools
```

---

**CurvTech Inc.** - Building the future of AI-powered education
