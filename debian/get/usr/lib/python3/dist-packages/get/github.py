import requests
from typing import Optional, Dict, Any

class GitHubClient:
    def __init__(self):
        self.session = requests.Session()
        self.base_url = "https://api.github.com"

    def get_release(self, repo: str, tag: str = "latest") -> Dict[str, Any]:
        """
        Fetches release information from GitHub.
        If tag is 'latest', fetches the latest release.
        Otherwise fetches the release by tag.
        """
        if tag == "latest":
            url = f"{self.base_url}/repos/{repo}/releases/latest"
        else:
            url = f"{self.base_url}/repos/{repo}/releases/tags/{tag}"

        response = self.session.get(url)
        if response.status_code == 404:
            raise ValueError(f"Release not found: {repo} @ {tag}")
        response.raise_for_status()
        return response.json()

    def get_tags(self, repo: str) -> list[Dict[str, Any]]:
        """Fetches list of tags for a repository."""
        url = f"{self.base_url}/repos/{repo}/tags"
        response = self.session.get(url)
        response.raise_for_status()
        return response.json()
