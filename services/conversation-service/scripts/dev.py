#!/usr/bin/env python3
"""
Development script for running the conversation service
使用 uv 运行开发环境的脚本
"""

import subprocess
import sys
import os
from pathlib import Path


def run_command(cmd: list[str], cwd: str | None = None) -> int:
    """运行命令并返回退出码"""
    print(f"🔧 运行命令: {' '.join(cmd)}")
    result = subprocess.run(cmd, cwd=cwd)
    return result.returncode


def main():
    """主函数"""
    service_dir = Path(__file__).parent.parent
    os.chdir(service_dir)

    if len(sys.argv) < 2:
        print("用法:")
        print("  python scripts/dev.py install   # 安装依赖")
        print("  python scripts/dev.py dev       # 安装开发依赖")
        print("  python scripts/dev.py run       # 运行服务")
        print("  python scripts/dev.py test      # 运行测试")
        print("  python scripts/dev.py lint      # 代码检查")
        print("  python scripts/dev.py format    # 代码格式化")
        sys.exit(1)

    command = sys.argv[1]

    if command == "install":
        # 安装生产依赖
        return run_command(["uv", "sync", "--no-dev"])

    elif command == "dev":
        # 安装所有依赖（包括开发依赖）
        return run_command(["uv", "sync"])

    elif command == "run":
        # 运行服务
        return run_command(["uv", "run", "python", "main.py"])

    elif command == "test":
        # 运行测试
        return run_command(["uv", "run", "pytest"])

    elif command == "test-cov":
        # 运行测试并生成覆盖率报告
        return run_command(["uv", "run", "pytest", "--cov"])

    elif command == "lint":
        # 代码检查
        print("🔍 运行 ruff 检查...")
        ruff_result = run_command(["uv", "run", "ruff", "check", "."])

        print("🔍 运行 mypy 类型检查...")
        mypy_result = run_command(["uv", "run", "mypy", "."])

        return max(ruff_result, mypy_result)

    elif command == "format":
        # 代码格式化
        print("🎨 运行 black 格式化...")
        black_result = run_command(["uv", "run", "black", "."])

        print("🎨 运行 ruff 自动修复...")
        ruff_result = run_command(["uv", "run", "ruff", "check", "--fix", "."])

        return max(black_result, ruff_result)

    elif command == "clean":
        # 清理缓存
        return run_command(["uv", "cache", "clean"])

    else:
        print(f"❌ 未知命令: {command}")
        return 1


if __name__ == "__main__":
    sys.exit(main())
