"""
Security and authentication utilities
"""

from fastapi import HTTPException, Depends
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from services.auth_service import AuthService
from config.settings import Settings

# Security
security = HTTPBearer()

# Global services (will be initialized in main.py)
auth_service: AuthService = None


def init_security(settings: Settings):
    """Initialize security services"""
    global auth_service
    auth_service = AuthService(settings)


async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
):
    """Validate JWT token and return user info"""
    try:
        user = await auth_service.validate_token(credentials.credentials)
        return user
    except Exception as e:
        raise HTTPException(
            status_code=401, detail="Invalid authentication credentials"
        )
