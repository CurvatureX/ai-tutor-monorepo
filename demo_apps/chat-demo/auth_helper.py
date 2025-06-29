#!/usr/bin/env python3
"""
Auth Helper for AI Tutor Chat Demo

This script helps generate JWT tokens for authentication with the conversation service.
It can also validate tokens and create demo users.
"""

import json
import sys
from datetime import datetime, timedelta
from typing import Dict, Any

try:
    import jwt
except ImportError:
    print("❌ PyJWT library not found. Please install it:")
    print("   pip install PyJWT")
    print("   or")
    print("   pip install -r requirements.txt")
    sys.exit(1)


class AuthHelper:
    """Helper class for JWT token operations"""

    def __init__(self, jwt_secret: str = "your-secret-key-change-in-production"):
        self.jwt_secret = jwt_secret

    def create_token(
        self, user_id: str, username: str, email: str, hours: int = 24
    ) -> str:
        """
        Create a JWT token for a user

        Args:
            user_id: Unique user identifier
            username: Username
            email: User email
            hours: Token validity in hours (default: 24)

        Returns:
            JWT token string
        """
        payload = {
            "user_id": user_id,
            "username": username,
            "email": email,
            "exp": datetime.utcnow() + timedelta(hours=hours),
            "iat": datetime.utcnow(),
        }

        return jwt.encode(payload, self.jwt_secret, algorithm="HS256")

    def validate_token(self, token: str) -> Dict[str, Any]:
        """
        Validate a JWT token and return payload

        Args:
            token: JWT token string

        Returns:
            Token payload dictionary

        Raises:
            jwt.ExpiredSignatureError: If token has expired
            jwt.InvalidTokenError: If token is invalid
        """
        try:
            payload = jwt.decode(token, self.jwt_secret, algorithms=["HS256"])

            # Check expiration
            if payload.get("exp", 0) < datetime.utcnow().timestamp():
                raise jwt.ExpiredSignatureError("Token has expired")

            return payload

        except jwt.ExpiredSignatureError:
            raise
        except jwt.InvalidTokenError:
            raise

    def create_demo_users(self) -> Dict[str, str]:
        """
        Create tokens for demo users

        Returns:
            Dictionary mapping usernames to tokens
        """
        demo_users = [
            {
                "user_id": "demo-user-001",
                "username": "alice_student",
                "email": "alice@demo.com",
            },
            {
                "user_id": "demo-user-002",
                "username": "bob_learner",
                "email": "bob@demo.com",
            },
            {
                "user_id": "demo-user-003",
                "username": "charlie_practice",
                "email": "charlie@demo.com",
            },
        ]

        tokens = {}
        for user in demo_users:
            token = self.create_token(user["user_id"], user["username"], user["email"])
            tokens[user["username"]] = token

        return tokens


def main():
    """Main CLI interface"""
    auth_helper = AuthHelper()

    if len(sys.argv) < 2:
        print("🔐 AI Tutor Auth Helper")
        print("=" * 40)
        print()
        print("Usage:")
        print("  python auth_helper.py create <user_id> <username> <email> [hours]")
        print("  python auth_helper.py validate <token>")
        print("  python auth_helper.py demo")
        print()
        print("Examples:")
        print("  python auth_helper.py create user123 john john@example.com")
        print("  python auth_helper.py create user456 mary mary@example.com 48")
        print(
            "  python auth_helper.py validate eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..."
        )
        print("  python auth_helper.py demo")
        print()
        print(
            "💡 Tip: Use 'dev-token' for development mode in the conversation service"
        )
        return

    command = sys.argv[1].lower()

    if command == "create":
        if len(sys.argv) < 5:
            print("❌ Error: create command requires user_id, username, and email")
            print(
                "Usage: python auth_helper.py create <user_id> <username> <email> [hours]"
            )
            return

        user_id = sys.argv[2]
        username = sys.argv[3]
        email = sys.argv[4]
        hours = int(sys.argv[5]) if len(sys.argv) > 5 else 24

        try:
            token = auth_helper.create_token(user_id, username, email, hours)
            print(f"✅ Token created successfully!")
            print(f"🆔 User ID: {user_id}")
            print(f"👤 Username: {username}")
            print(f"📧 Email: {email}")
            print(f"⏰ Valid for: {hours} hours")
            print()
            print(f"🔑 JWT Token:")
            print(token)
            print()
            print("💡 Copy this token and use it in the chat demo's 'Auth Token' field")

        except Exception as e:
            print(f"❌ Error creating token: {e}")

    elif command == "validate":
        if len(sys.argv) < 3:
            print("❌ Error: validate command requires a token")
            print("Usage: python auth_helper.py validate <token>")
            return

        token = sys.argv[2]

        try:
            payload = auth_helper.validate_token(token)
            print("✅ Token is valid!")
            print(f"🆔 User ID: {payload.get('user_id')}")
            print(f"👤 Username: {payload.get('username')}")
            print(f"📧 Email: {payload.get('email')}")
            print(f"⏰ Expires: {datetime.fromtimestamp(payload.get('exp', 0))}")
            print(f"🕐 Issued: {datetime.fromtimestamp(payload.get('iat', 0))}")

        except jwt.ExpiredSignatureError:
            print("❌ Token has expired")
        except jwt.InvalidTokenError:
            print("❌ Token is invalid")
        except Exception as e:
            print(f"❌ Error validating token: {e}")

    elif command == "demo":
        print("🎭 Creating demo user tokens...")
        print("=" * 40)

        try:
            demo_tokens = auth_helper.create_demo_users()

            for username, token in demo_tokens.items():
                print(f"\n👤 {username}")
                print(f"🔑 {token}")

            print("\n" + "=" * 40)
            print("✅ Demo tokens created successfully!")
            print()
            print("💡 Usage Tips:")
            print("  • Copy any token above and paste it in the chat demo")
            print("  • Tokens are valid for 24 hours")
            print("  • For development, you can also use 'dev-token'")
            print("  • Each token represents a different demo user")

        except Exception as e:
            print(f"❌ Error creating demo tokens: {e}")

    else:
        print(f"❌ Unknown command: {command}")
        print("Valid commands: create, validate, demo")


if __name__ == "__main__":
    main()
