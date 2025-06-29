"""
Language learning session endpoints
"""

from fastapi import APIRouter, HTTPException
from app.api.deps import CurrentUser, AIServiceDep
from services.ai_service import AIService

router = APIRouter()


@router.get("/language/session")
async def get_user_language_session(
    user=CurrentUser,
    ai_service: AIService = AIServiceDep
):
    """Get user's language learning session and conversation history"""
    try:
        user_id = user.get("user_id")

        # 获取用户的所有sessions
        sessions = await ai_service.get_user_sessions(user_id)

        # 如果用户没有session，创建一个默认的英语口语学习session
        if not sessions:
            session = await ai_service.create_tutor_session(
                user_id=user_id, subject="English", level="intermediate"
            )
            return {
                "session": {
                    "id": session.id,
                    "user_id": session.user_id,
                    "language": "English",
                    "level": session.level,
                    "goals": ["Daily conversation", "Pronunciation improvement"],
                    "created_at": session.created_at,
                    "last_activity": session.last_activity,
                    "message_count": session.message_count,
                    "status": session.status,
                },
                "history": [],
                "learning_context": {
                    "today_focus": "General speaking practice",
                    "session_type": "conversation",
                    "emotional_state": "ready_to_learn",
                },
            }

        # 使用第一个session作为用户的唯一session
        session = sessions[0]
        history = await ai_service.get_session_history(session.id, user_id)

        return {
            "session": {
                "id": session.id,
                "user_id": session.user_id,
                "language": getattr(session, "subject", "English"),  # 兼容旧数据
                "level": session.level,
                "goals": getattr(session, "goals", ["Daily conversation"]),
                "created_at": session.created_at,
                "last_activity": session.last_activity,
                "message_count": session.message_count,
                "status": session.status,
            },
            "history": history,
            "learning_context": {
                "today_focus": "Speaking practice",
                "session_type": "conversation",
                "emotional_state": "engaged",
            },
        }
    except Exception as e:
        raise HTTPException(
            status_code=500, detail=f"Error fetching language session: {str(e)}"
        )
