#!/usr/bin/env python3
"""
Setup script for AI Tutor Chat Demo

This script helps users set up the demo environment quickly.
"""

import subprocess
import sys
import os
import webbrowser
from pathlib import Path


def check_python_version():
    """Check if Python version is compatible"""
    if sys.version_info < (3, 7):
        print("❌ Python 3.7 or higher is required")
        print(f"   Your version: {sys.version}")
        return False
    print(f"✅ Python version: {sys.version.split()[0]}")
    return True


def install_dependencies():
    """Install required dependencies"""
    print("📦 Installing dependencies...")
    try:
        subprocess.check_call(
            [sys.executable, "-m", "pip", "install", "-r", "requirements.txt"]
        )
        print("✅ Dependencies installed successfully")
        return True
    except subprocess.CalledProcessError as e:
        print(f"❌ Failed to install dependencies: {e}")
        print("💡 Try running manually: pip install PyJWT")
        return False


def check_conversation_service():
    """Check if conversation service is running"""
    import urllib.request
    import urllib.error

    try:
        with urllib.request.urlopen(
            "http://localhost:8000/docs", timeout=5
        ) as response:
            if response.getcode() == 200:
                print("✅ Conversation service is running")
                return True
    except (urllib.error.URLError, urllib.error.HTTPError):
        pass

    print("⚠️  Conversation service not detected on localhost:8000")
    print("💡 Start it with: cd ../../services/conversation-service && python main.py")
    return False


def create_demo_token():
    """Create a demo token for testing"""
    try:
        from auth_helper import AuthHelper

        auth_helper = AuthHelper()
        token = auth_helper.create_token(
            "demo-user-001", "demo_user", "demo@example.com"
        )

        print("🔑 Demo token created:")
        print(f"   {token}")
        print()
        print("💡 You can use this token in the chat demo")
        return token
    except Exception as e:
        print(f"⚠️  Could not create demo token: {e}")
        print("💡 You can use 'dev-token' for development mode")
        return "dev-token"


def start_demo_server():
    """Start the demo server"""
    try:
        print("🚀 Starting demo server...")
        print("💡 Press Ctrl+C to stop the server")
        print("🌐 Demo will open automatically in your browser")
        print()

        import server

        server.main()
    except KeyboardInterrupt:
        print("\n👋 Demo server stopped")
    except Exception as e:
        print(f"❌ Failed to start server: {e}")
        print("💡 You can also open index.html directly in your browser")


def main():
    """Main setup process"""
    print("🤖 AI Tutor Chat Demo Setup")
    print("=" * 40)
    print()

    # Check Python version
    if not check_python_version():
        return

    # Install dependencies
    if not install_dependencies():
        return

    # Check conversation service
    service_running = check_conversation_service()

    # Create demo token
    demo_token = create_demo_token()

    print()
    print("🎉 Setup complete!")
    print("=" * 40)

    if not service_running:
        print("⚠️  Warning: Conversation service is not running")
        print("📝 To start the service:")
        print("   1. Open a new terminal")
        print("   2. cd ../../services/conversation-service")
        print("   3. python main.py")
        print()

    print("🚀 Demo Options:")
    print("   1. Start demo server: python server.py")
    print("   2. Open index.html directly in your browser")
    print("   3. Generate auth tokens: python auth_helper.py demo")
    print()

    choice = input("Start demo server now? (y/n): ").lower().strip()
    if choice in ["y", "yes"]:
        start_demo_server()
    else:
        print("👍 Run 'python server.py' when you're ready to start the demo")


if __name__ == "__main__":
    main()
