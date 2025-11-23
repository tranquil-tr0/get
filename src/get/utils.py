import os
import platform
import re
from pathlib import Path
from typing import Tuple

def get_platform_info() -> Tuple[str, str]:
    """Returns the OS and architecture."""
    system = platform.system().lower()
    machine = platform.machine().lower()
    
    # Normalize architecture
    if machine in ["x86_64", "amd64"]:
        machine = "amd64"
    elif machine in ["aarch64", "arm64"]:
        machine = "arm64"
        
    return system, machine

def parse_repo_url(url: str) -> str:
    """
    Parses a GitHub repository URL or 'user/repo' string and returns 'user/repo'.
    """
    url = url.strip()
    match = re.match(r"^(?:https?://github\.com/)?([^/]+/[^/]+)(?:\.git)?$", url)
    if match:
        return match.group(1)
    raise ValueError(f"Invalid repository URL or identifier: {url}")

def get_data_dir() -> Path:
    """Returns the data directory for the application."""
    # Strictly ~/.local/share/get for Linux
    return Path.home() / ".local" / "share" / "get"

def ensure_dir(path: Path):
    """Ensures a directory exists."""
    path.mkdir(parents=True, exist_ok=True)
