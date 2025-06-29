#!/usr/bin/env python3
"""
Simple HTTP Server for AI Tutor Chat Demo

Optional local server to serve the chat demo files.
Useful for development and avoiding CORS issues with file:// protocol.
"""

import http.server
import socketserver
import os
import sys
import webbrowser
from pathlib import Path


class ChatDemoHandler(http.server.SimpleHTTPRequestHandler):
    """Custom handler for the chat demo"""

    def __init__(self, *args, **kwargs):
        # Set the directory to serve files from
        super().__init__(*args, directory=str(Path(__file__).parent), **kwargs)

    def end_headers(self):
        # Add CORS headers for local development
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        super().end_headers()

    def do_OPTIONS(self):
        # Handle preflight CORS requests
        self.send_response(200)
        self.end_headers()

    def log_message(self, format, *args):
        # Custom log format
        print(f"🌐 {self.address_string()} - {format % args}")


def main():
    """Start the demo server"""
    port = 3000

    # Check if port is specified
    if len(sys.argv) > 1:
        try:
            port = int(sys.argv[1])
        except ValueError:
            print("❌ Invalid port number. Using default port 3000.")

    # Check if index.html exists
    if not os.path.exists("index.html"):
        print("❌ index.html not found in current directory")
        print(
            "💡 Make sure you're running this script from the demo_apps/chat-demo directory"
        )
        return

    try:
        # Create server
        with socketserver.TCPServer(("", port), ChatDemoHandler) as httpd:
            print("🚀 AI Tutor Chat Demo Server")
            print("=" * 40)
            print(f"📍 Serving at: http://localhost:{port}")
            print(f"📁 Directory: {os.getcwd()}")
            print(f"🔗 Chat Demo: http://localhost:{port}/index.html")
            print()
            print("💡 Tips:")
            print("  • Make sure the conversation service is running on port 8000")
            print("  • Use Ctrl+C to stop the server")
            print("  • Check browser console for any errors")
            print()

            # Try to open browser automatically
            try:
                webbrowser.open(f"http://localhost:{port}")
                print("🌐 Opening browser automatically...")
            except Exception:
                print("🌐 Please open http://localhost:{port} in your browser")

            print("📡 Server starting...")
            httpd.serve_forever()

    except OSError as e:
        if e.errno == 48:  # Address already in use
            print(f"❌ Port {port} is already in use")
            print(f"💡 Try a different port: python server.py {port + 1}")
        else:
            print(f"❌ Server error: {e}")
    except KeyboardInterrupt:
        print("\n👋 Server stopped by user")
    except Exception as e:
        print(f"❌ Unexpected error: {e}")


if __name__ == "__main__":
    main()
