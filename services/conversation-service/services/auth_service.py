import jwt
import httpx
from datetime import datetime, timedelta
from typing import Dict, Any, Optional
from config.settings import Settings


class AuthService:
    """Authentication and authorization service"""

    def __init__(self, settings: Settings):
        self.settings = settings
        self.user_service_url = settings.user_service_url

    async def health_check(self) -> bool:
        """Check if auth service is healthy"""
        try:
            # In production, check connection to user service
            return True
        except Exception:
            return False

    async def validate_token(self, token: str) -> Dict[str, Any]:
        """Validate JWT token and return user info"""

        try:
            # In development mode, allow bypass
            if self.settings.debug and token == "dev-token":
                return {
                    "user_id": "dev-user-123",
                    "username": "dev_user",
                    "email": "dev@example.com",
                }

            # Decode JWT token
            payload = jwt.decode(token, self.settings.jwt_secret, algorithms=["HS256"])

            # Check expiration
            if payload.get("exp", 0) < datetime.utcnow().timestamp():
                raise jwt.ExpiredSignatureError("Token has expired")

            # Verify with user service (optional)
            user_info = await self._verify_with_user_service(payload.get("user_id"))

            return {
                "user_id": payload.get("user_id"),
                "username": payload.get("username"),
                "email": payload.get("email"),
                **user_info,
            }

        except jwt.ExpiredSignatureError:
            raise Exception("Token has expired")
        except jwt.InvalidTokenError:
            raise Exception("Invalid token")
        except Exception as e:
            raise Exception(f"Token validation failed: {str(e)}")

    async def _verify_with_user_service(self, user_id: str) -> Dict[str, Any]:
        """Verify user with user service"""

        try:
            if not self.user_service_url:
                return {}

            async with httpx.AsyncClient() as client:
                response = await client.get(
                    f"{self.user_service_url}/users/{user_id}", timeout=5.0
                )

                if response.status_code == 200:
                    return response.json()
                else:
                    return {}

        except Exception as e:
            print(f"Warning: Could not verify with user service: {e}")
            return {}

    def create_token(self, user_id: str, username: str, email: str) -> str:
        """Create JWT token for user (for testing purposes)"""

        payload = {
            "user_id": user_id,
            "username": username,
            "email": email,
            "exp": datetime.utcnow() + timedelta(hours=24),
            "iat": datetime.utcnow(),
        }

        return jwt.encode(payload, self.settings.jwt_secret, algorithm="HS256")
