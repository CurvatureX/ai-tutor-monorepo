---
sidebar_position: 2
title: Conversation 模块技术设计
description: 具备长期记忆、学习目标意识和实时用户画像更新的 AI 口语学习助手技术设计文档
keywords: [AI, 口语学习, 对话, RAG, 向量数据库, LLM]
---

# AI Conversation 模块技术设计

## 1. 项目概述 (Overview)

### 目标和范围

- **核心目标**: 构建具备长期记忆、学习目标意识和实时用户画像更新的 AI 口语学习助手，通过个性化对话持续提升用户英语水平
- **AI Agent 能力**:
  - 长期记忆和上下文理解
  - 学习进度跟踪和目标引导
  - 个性化语调和话题适配
  - 情绪感知和适应性反馈
  - 发散对话的学习计划回归
- **用户场景**: 英语口语练习、日常对话陪伴、学习进度跟踪、个性化教学
- **成功指标**:
  - 对话连贯性和上下文准确率 >95%
  - 学习目标完成率 >80%
  - 用户满意度 >4.5/5
  - 平均会话时长 >15 分钟

### 功能边界

- **包含功能**:

  - 用户画像建模和动态更新
  - 学习进度跟踪和计划制定
  - 长期记忆管理和检索
  - 情绪感知和个性化响应
  - 口语练习引导和纠错
  - 多轮对话上下文管理

- **不包含功能**:

  - 语音识别和合成（由 speech-service 处理）
  - 学习内容生成（依赖外部课程系统）
  - 用户认证和权限管理（由 user-service 处理）

- **集成点**:
  - Speech Service: 语音转文本和文本转语音
  - User Service: 用户基础信息和认证
  - Analytics Service: 学习数据分析和报告

---

## 2. AI 模型设计 (AI Model Design)

### 模型架构

- **基础模型**: GPT-4o 或 Claude-3.5-Sonnet 作为主对话模型
- **模型微调**: 针对教学场景的 Few-shot learning，无需完整微调
- **多模态支持**: 主要处理文本对话，集成语音转文本结果

### Prompt 工程

- **系统提示词**: 核心的教学助手人格设定

```markdown
# 核心 System Prompt

你是一位专业的英语口语学习助手，具备以下特征：

## 角色定位

- 友好、耐心、鼓励性的英语学习伙伴
- 具备长期记忆，能记住用户的学习历程和个人喜好
- 目标导向，始终关注用户的学习进展和目标达成

## 核心能力

1. **个性化教学**: 根据用户画像调整语调、难度和话题
2. **进度跟踪**: 持续关注学习目标，适时引导回归学习计划
3. **情绪感知**: 识别用户情绪状态，调整鼓励方式
4. **记忆连贯**: 引用历史对话内容，保持长期关系

## 输出格式

- 使用自然对话语调
- 重点词汇用 **加粗** 标记
- 提供具体的语言学习建议
- 适时插入相关的历史学习内容
```

- **动态 Prompt 构建**: 基于用户画像和学习上下文的动态提示词生成
- **RAG 增强**: 结合检索到的历史对话和学习记录

### 动态 Prompt 构建架构

#### 架构设计原则

采用 **混合策略**：**固定核心 System Prompt + 动态上下文注入**

- **固定核心 System Prompt**: 保持 AI 助手的核心人格和角色定位稳定
- **动态上下文注入**: 每次对话时注入个性化的用户画像、学习状态和相关记忆

#### Prompt 构建流程

```python
class ConversationPromptBuilder:
    def __init__(self):
        self.core_system_prompt = CORE_SYSTEM_PROMPT

    async def build_messages(self, user_input: str, user_id: str):
        # 1. 获取用户画像
        user_profile = await self.context_manager.get_user_profile(user_id)

        # 2. 获取学习上下文
        learning_context = await self.context_manager.get_learning_context(user_id)

        # 3. RAG 检索相关记忆
        rag_memories = await self.rag_engine.retrieve_memories(
            user_input, user_id, top_k=3
        )

        # 4. 构建完整 messages
        messages = [
            {
                "role": "system",
                "content": self.core_system_prompt
            },
            {
                "role": "user",
                "content": self.build_context_message(
                    user_profile, learning_context, rag_memories
                )
            },
            {
                "role": "user",
                "content": user_input
            }
        ]

        return messages
```

#### 动态上下文模板

```python
def build_context_message(user_profile, learning_context, rag_memories):
    context_message = f"""
## 📋 当前用户信息
**姓名**: {user_profile['name']}
**英语水平**: {user_profile['english_level']}
**学习目标**: {', '.join(user_profile['goals'])}
**性格特点**: {user_profile['personality']}
**偏好语调**: {user_profile['preferred_tone']}
**学习难点**: {', '.join(user_profile['speaking_challenges'])}

## 🎯 今日学习计划
**主题**: {learning_context['today_plan']['topic']}
**目标词汇**: {', '.join(learning_context['today_plan']['target_vocab'])}
**语法重点**: {learning_context['today_plan']['grammar_focus']}
**情绪状态**: {learning_context['emotional_state']}

## 🧠 相关记忆
{format_rag_memories(rag_memories)}

---
请基于以上信息与用户进行个性化对话。
"""
    return context_message
```

#### 分层上下文管理

| 更新频率                 | 上下文类型 | 包含内容                          | 用途           |
| ------------------------ | ---------- | --------------------------------- | -------------- |
| **高频**（每轮对话）     | 短期上下文 | 最近 5 条消息、当前情绪、即时反馈 | 保持对话连贯性 |
| **中频**（每日/每课）    | 中期上下文 | 今日计划、课程摘要、进度更新      | 学习目标引导   |
| **低频**（用户画像变化） | 长期上下文 | 用户画像、学习目标、交流偏好      | 个性化基础     |

#### Token 预算管理

```python
class TokenBudgetManager:
    def __init__(self, max_tokens=4000):
        self.max_tokens = max_tokens
        self.core_prompt_tokens = 500  # 固定部分
        self.reserved_tokens = 1000    # 预留给响应

    def optimize_context(self, context_parts):
        available_tokens = self.max_tokens - self.core_prompt_tokens - self.reserved_tokens

        # 按优先级排序上下文
        prioritized_context = self.prioritize_context(context_parts)

        # 逐步添加直到达到token限制
        final_context = self.fit_within_budget(prioritized_context, available_tokens)

        return final_context

    def prioritize_context(self, context_parts):
        """上下文优先级排序"""
        priority_order = [
            'user_profile',      # 优先级1：基础用户信息
            'learning_goals',    # 优先级2：学习目标
            'recent_mistakes',   # 优先级3：近期错误记录
            'session_context',   # 优先级4：当前会话上下文
            'historical_memories' # 优先级5：历史记忆
        ]
        return sorted(context_parts, key=lambda x: priority_order.index(x['type']))
```

#### 智能上下文过滤

```python
def filter_context_by_relevance(context_data, user_input, threshold=0.7):
    """只包含与当前对话相关的上下文信息"""
    relevant_context = {}

    # 使用语义相似度过滤
    for key, value in context_data.items():
        similarity = calculate_similarity(user_input, str(value))
        if similarity > threshold:
            relevant_context[key] = value

    return relevant_context
```

#### 实施阶段规划

| 阶段        | 实现内容                              | 预期效果           | 技术重点         |
| ----------- | ------------------------------------- | ------------------ | ---------------- |
| **Phase 1** | 固定 System Prompt + 基础用户画像注入 | 个性化对话基础能力 | 模板化上下文构建 |
| **Phase 2** | 添加 RAG 记忆检索和学习上下文         | 长期记忆和学习引导 | 向量检索集成     |
| **Phase 3** | 智能上下文过滤和 Token 优化           | 性能优化和成本控制 | 算法优化         |
| **Phase 4** | 动态 Prompt 模板和 A/B 测试           | 效果优化和持续改进 | 数据驱动优化     |

### 知识库和检索

- **知识源**:

  - 用户历史对话记录
  - 学习进度和反馈数据
  - 个性化用户画像
  - 英语学习资源库

- **向量数据库**: Weaviate（支持混合结构化+嵌入检索）
- **检索策略**:

  - 语义相似度检索
  - 时间衰减权重
  - 内容类型优先级
  - 错误记录高权重保留

- **知识更新**: 每次对话后实时更新记忆库，每日生成学习摘要

---

## 3. 系统架构 (System Architecture)

### 整体架构

```
用户界面 → Gateway → Conversation Service → LLM API
                         ↓
                    Memory Manager
                         ↓
                  Weaviate向量数据库
                         ↓
                  User/Learning Context
                         ↓
                Speech/User/Analytics Services
```

### 核心组件

- **Conversation Controller**: 主对话流程管理和路由
- **Context Manager**: 用户画像和学习上下文管理
- **Memory Manager**: 长期记忆存储、检索和更新
- **RAG Engine**: 记忆检索和 Prompt 增强
- **Learning Tracker**: 学习进度跟踪和目标管理
- **Emotion Analyzer**: 用户情绪识别和分析

### 技术选型

- **后端框架**: Python FastAPI
- **LLM 接口**: OpenAI API (GPT-4o) / Anthropic API (Claude-3.5)
- **向量数据库**: Weaviate
- **缓存**: Redis (会话状态和频繁访问数据)
- **消息队列**: RabbitMQ (异步记忆处理)
- **数据库**: PostgreSQL (结构化用户数据)

---

## 4. 数据流和交互 (Data Flow & Interactions)

### 典型对话流程

1. **用户输入** → 接收文本/语音转文本结果
2. **上下文加载** → 获取用户画像和当前学习状态
3. **记忆检索** → RAG 检索相关历史对话和学习记录
4. **情绪分析** → 分析用户当前情绪状态
5. **Prompt 构建** → 结合所有上下文生成完整 prompt
6. **LLM 调用** → 获取 AI 响应
7. **记忆更新** → 更新对话记录和学习状态
8. **响应返回** → 格式化输出给用户

### 数据结构设计

#### User Profile Context

```json
{
  "user_id": "u001",
  "name": "Alex",
  "age": 25,
  "gender": "male",
  "native_language": "Chinese",
  "english_level": "B1",
  "goals": ["IELTS speaking 7", "Confident in business meetings"],
  "interests": ["travel", "technology", "movies"],
  "personality": "introvert, likes deep thinking",
  "preferred_tone": "friendly and encouraging",
  "communication_style": "prefers text",
  "speaking_challenges": ["pronunciation", "fluency under pressure"],
  "learning_style": "task-based",
  "emotion_preference": "喜欢被鼓励",
  "cultural_sensitivity": "关注英式美式差异和用词禁忌",
  "pressure": "轻松学习状态",
  "speed_preference": "慢速",
  "created_at": "2025-06-01",
  "updated_at": "2025-06-28"
}
```

#### Learning Context

```json
{
  "user_id": "u001",
  "date": "2025-06-28",
  "today_plan": {
    "topic": "Self Introduction",
    "target_vocab": ["curious", "ambitious"],
    "grammar_focus": "present simple",
    "mode": "speaking practice"
  },
  "history_summary": "Reviewed vocabulary on emotions. Practiced giving opinions about movies.",
  "completed": true,
  "score": {
    "fluency": 6,
    "accuracy": 7,
    "engagement": 8
  },
  "notes": "User had difficulty pronouncing 'curious', showed enthusiasm when discussing travel.",
  "emotional_state": "motivated",
  "followup_recommendation": "Revise emotion vocabulary. Introduce question forms tomorrow.",
  "recording_url": "s3://.../session_20250628.mp3"
}
```

#### Memory Vector Document

```json
{
  "user_id": "u001",
  "doc_id": "20250628_movie_session",
  "content_type": "session_summary",
  "text": "User talked about their favorite movies. Struggled with expressing opinions clearly.",
  "embedding": [0.021, 0.113, ...],
  "tags": ["movie", "opinion", "past_tense"],
  "timestamp": "2025-06-28T14:00:00Z",
  "importance_score": 0.8
}
```

### 外部集成

- **Speech Service**: 语音转文本结果接收，TTS 请求发送
- **User Service**: 用户基础信息同步
- **Analytics Service**: 学习数据上报和分析结果获取

---

## 5. 性能和可靠性 (Performance & Reliability)

### 性能指标

- **响应时间**: 端到端延迟 < 3 秒（包含 LLM 调用）
- **并发处理**: 支持 1000+ 并发用户
- **Token 消耗**: 平均每次对话 < 2000 tokens
- **缓存命中率**: 用户上下文缓存命中率 > 90%

### 可靠性保障

- **错误处理**:

  - LLM API 调用失败重试机制（3 次，指数退避）
  - 向量数据库查询超时处理（5 秒超时）
  - 降级到简化 prompt 继续服务

- **降级策略**:

  - LLM 不可用时使用预定义回复模板
  - 向量检索失败时使用基础用户画像
  - 记忆更新失败不影响当前对话

- **监控告警**:
  - API 调用成功率 < 95% 告警
  - 平均响应时间 > 5 秒 告警
  - 向量数据库连接状态监控

### 安全和合规

- **输入验证**:

  - 用户输入长度限制（< 1000 字符）
  - 敏感词过滤和内容安全检查
  - SQL 注入和 Prompt 注入防护

- **输出过滤**:

  - LLM 输出内容安全检查
  - 个人隐私信息过滤
  - 不当内容识别和拦截

- **数据隐私**:
  - 对话数据加密存储
  - 用户画像数据访问控制
  - 符合 GDPR 数据保护要求

---

## 6. 测试和评估 (Testing & Evaluation)

### AI 模型测试

- **Prompt 测试**:

  - 不同用户画像下的响应质量
  - 记忆检索准确性验证
  - 学习目标引导效果测试

- **准确性测试**:

  - 上下文理解准确率测试
  - 历史记忆引用正确性
  - 学习进度跟踪准确性

- **边界测试**:
  - 超长对话上下文处理
  - 多语言混合输入处理
  - 异常用户行为应对

### 系统测试

- **性能测试**:

  - 1000 并发用户负载测试
  - 向量检索响应时间测试
  - 内存使用和 CPU 性能测试

- **集成测试**:
  - 与 Speech Service 集成验证
  - 用户数据同步测试
  - 跨服务数据一致性验证

### 评估指标

- **技术指标**:

  - 对话响应时间 < 2 秒
  - 记忆检索准确率 > 90%
  - 系统可用性 > 99.5%

- **业务指标**:
  - 用户学习目标完成率
  - 对话质量评分
  - 用户留存率和活跃度

---

## 7. 部署和运维 (Deployment & Operations)

### 部署策略

- **环境配置**:

  - 开发环境：单机部署，使用 Chroma 向量数据库
  - 测试环境：小规模集群，模拟生产配置
  - 生产环境：Kubernetes 集群，Weaviate 集群

- **发布流程**:
  - GitLab CI/CD 自动化部署
  - 蓝绿部署确保零停机
  - Prompt 版本管理和 A/B 测试

### 运维监控

- **核心指标**:

  - LLM API 调用成功率和延迟
  - 向量数据库查询性能
  - 用户会话活跃状态

- **业务指标**:

  - 每日活跃对话数
  - 学习目标完成情况
  - 用户满意度反馈

- **成本监控**:
  - LLM Token 消耗量和费用
  - 向量数据库存储成本
  - 服务器资源使用情况

### 持续优化

- **Prompt 迭代**:

  - 基于用户反馈优化 prompt 模板
  - A/B 测试不同版本的效果
  - 定期更新教学策略

- **记忆库维护**:
  - 定期清理过期或低价值记忆
  - 重要学习记录权重调整
  - 向量索引优化和重建

---

## 附录

### 风险和缓解

| 风险项           | 影响 | 概率 | 缓解措施                   |
| ---------------- | ---- | ---- | -------------------------- |
| LLM API 不稳定   | 高   | 中   | 多厂商备选，本地模型备份   |
| 向量检索延迟     | 中   | 低   | 缓存优化，查询超时处理     |
| 用户隐私泄露     | 高   | 低   | 数据加密，访问控制         |
| 记忆数据丢失     | 中   | 低   | 定期备份，多副本存储       |
| Token 成本超预算 | 中   | 中   | 使用量监控，Prompt 优化    |
| 对话质量下降     | 高   | 中   | 质量监控，及时 Prompt 调优 |

### RAG 检索策略详解

- **时间衰减**: 超过 7 天的记忆权重衰减 50%，超过 30 天衰减 80%
- **内容标签匹配**: 相同主题标签权重 +0.3，错误记录权重 +0.5
- **检索优化**: Top-K=5，相似度阈值 0.7，混合检索（语义+关键词）
- **Prompt 拼接**: 最多插入 3 条历史记录，总长度控制在 500 tokens 内
