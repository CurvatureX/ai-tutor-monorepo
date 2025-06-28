---
sidebar_position: 1
title: API Documentation Overview
description: AI Tutor 项目的 API 接口文档概览
keywords: [API, gRPC, REST, 接口文档]
---

# API Documentation Overview

This section contains all API documentation for the AI Tutor platform.

## 📋 API 文档结构

### 核心服务 API

- **User Service API**: 用户管理、认证、配置接口
- **Conversation Service API**: 对话管理、AI 交互接口
- **Speech Service API**: 语音识别、合成、评估接口
- **Analytics Service API**: 数据分析、报告接口
- **Notification Service API**: 消息推送、通知接口

### 协议规范

- **gRPC Protocols**: Protocol Buffers 定义和接口规范
- **REST APIs**: HTTP 接口和数据格式
- **WebSocket**: 实时通信协议

### 客户端集成

- **Flutter Integration**: 移动端集成指南
- **Unity Integration**: 3D 应用集成指南
- **Web Integration**: 网页端集成指南

## 🚀 快速开始

1. **协议定义**: 查看 `shared/proto/` 目录下的 Protocol Buffers 定义
2. **代码生成**: 使用 `tools/proto-gen/` 工具生成客户端代码
3. **API 测试**: 使用提供的测试工具验证接口

## 📖 更多信息

- [技术架构设计](/tech_design/架构设计/)
- [Conversation 模块设计](/tech_design/conversation_core_tech_design/)
- [开发提示与技巧](/prompt_cheatsheet/cursor_prompt/)
