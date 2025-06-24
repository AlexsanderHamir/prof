import json
import urllib.request
from typing import Optional, Tuple

CURRENT_VERSION = "1.0.25"

GITHUB_REPO = "AlexsanderHamir/prof"
GITHUB_API_URL = f"https://api.github.com/repos/{GITHUB_REPO}/releases/latest"


def normalize_version(version: str) -> str:
    return version.lstrip('v') if version.startswith('v') else version


def get_latest_version() -> Optional[str]:
    try:
        with urllib.request.urlopen(GITHUB_API_URL) as response:
            data = json.loads(response.read().decode())
            return data.get('tag_name')
    except Exception:
        return None


def check_version() -> Tuple[str, Optional[str]]:
    latest_version = get_latest_version()
    return CURRENT_VERSION, latest_version


def format_version_output(current_version: str, latest_version: Optional[str]) -> str:
    output = f"Current version: {current_version}"

    if latest_version:
        normalized_current = normalize_version(current_version)
        normalized_latest = normalize_version(latest_version)

        if normalized_latest == normalized_current:
            output += f"\nLatest version: {latest_version} (up to date)"
        else:
            output += f"\nLatest version: {latest_version} (update available)"
    else:
        output += "\nLatest version: Unable to fetch (check your internet connection)"

    return output
