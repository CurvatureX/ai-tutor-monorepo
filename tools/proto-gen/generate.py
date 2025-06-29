#!/usr/bin/env python3
"""
Protocol Buffer code generation script for ai-tutor-monorepo
"""

import os
import subprocess
import sys
from pathlib import Path

def run_command(cmd, cwd=None):
    """Run a shell command and return the result"""
    try:
        result = subprocess.run(cmd, shell=True, cwd=cwd, capture_output=True, text=True)
        if result.returncode != 0:
            print(f"Error running command: {cmd}")
            print(f"stdout: {result.stdout}")
            print(f"stderr: {result.stderr}")
            return False
        return True
    except Exception as e:
        print(f"Exception running command: {cmd}, error: {e}")
        return False

def generate_go_protos():
    """Generate Go code from proto files"""
    project_root = Path(__file__).parent.parent.parent
    proto_dir = project_root / "shared" / "proto"
    
    # Create output directories
    output_dirs = [
        project_root / "gateway" / "pkg" / "proto",
        project_root / "services" / "speech-service" / "pkg" / "proto"
    ]
    
    for output_dir in output_dirs:
        output_dir.mkdir(parents=True, exist_ok=True)
    
    # Find all proto files
    proto_files = []
    for proto_file in proto_dir.rglob("*.proto"):
        proto_files.append(str(proto_file))
    
    if not proto_files:
        print("No proto files found")
        return False
    
    print(f"Found {len(proto_files)} proto files")
    
    # Generate Go code for each service
    for output_dir in output_dirs:
        print(f"Generating Go code to {output_dir}")
        
        cmd = f"""protoc \
            --proto_path={proto_dir} \
            --go_out={output_dir} \
            --go_opt=paths=source_relative \
            --go-grpc_out={output_dir} \
            --go-grpc_opt=paths=source_relative \
            {' '.join(proto_files)}"""
        
        if not run_command(cmd, cwd=project_root):
            return False
    
    return True

def main():
    """Main function"""
    print("Generating protocol buffer code...")
    
    # Check if protoc is installed
    if not run_command("protoc --version"):
        print("protoc is not installed. Please install Protocol Buffers compiler.")
        sys.exit(1)
    
    # Check if protoc-gen-go is installed
    if not run_command("protoc-gen-go --version"):
        print("protoc-gen-go is not installed. Installing...")
        if not run_command("go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"):
            print("Failed to install protoc-gen-go")
            sys.exit(1)
    
    # Check if protoc-gen-go-grpc is installed
    if not run_command("protoc-gen-go-grpc --version"):
        print("protoc-gen-go-grpc is not installed. Installing...")
        if not run_command("go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"):
            print("Failed to install protoc-gen-go-grpc")
            sys.exit(1)
    
    # Generate Go code
    if generate_go_protos():
        print("Successfully generated protocol buffer code")
    else:
        print("Failed to generate protocol buffer code")
        sys.exit(1)

if __name__ == "__main__":
    main()