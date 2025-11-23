import json
import os
import shutil
import subprocess
import tempfile
import zipfile
import tarfile
from pathlib import Path
from typing import Dict, Any, Optional, List, Tuple
from datetime import datetime

from .utils import get_data_dir, ensure_dir, get_platform_info
from .github import GitHubClient

class PackageManager:
    def __init__(self, metadata_path: Optional[Path] = None):
        if metadata_path:
            self.metadata_path = metadata_path
        else:
            self.metadata_path = get_data_dir() / "get.json"
        
        self.client = GitHubClient()
        self.installed_packages: Dict[str, Dict[str, Any]] = {}
        self._load_metadata()

    def _load_metadata(self):
        if self.metadata_path.exists():
            try:
                with open(self.metadata_path, "r") as f:
                    self.installed_packages = json.load(f)
            except json.JSONDecodeError:
                self.installed_packages = {}
        else:
            self.installed_packages = {}

    def _save_metadata(self):
        ensure_dir(self.metadata_path.parent)
        with open(self.metadata_path, "w") as f:
            json.dump(self.installed_packages, f, indent=4)

    def install(self, repo: str, release_tag: str = "latest", options: Optional[Dict] = None):
        """Installs a package from a GitHub repository."""
        print(f"Fetching release for {repo} ({release_tag})...")
        release = self.client.get_release(repo, release_tag)
        tag_name = release["tag_name"]
        
        # Asset selection logic
        asset = self._select_best_asset(release["assets"])
        if not asset:
            raise ValueError("No suitable asset found for your platform.")

        print(f"Downloading {asset['name']}...")
        download_path = self._download_asset(asset["browser_download_url"], asset["name"])
        
        print(f"Installing...")
        try:
            install_info = self._install_asset(download_path, repo, release, options)
            
            # Update metadata
            self.installed_packages[repo] = {
                "version": tag_name.lstrip("v"),
                "installed_at": datetime.now().isoformat(),
                "original_name": asset["name"],
                "install_type": install_info["type"],
                "apt_name": install_info.get("apt_name"),
                "binary_path": install_info.get("binary_path"),
                "tag_prefix": options.get("tag_prefix", "") if options else ""
            }
            self._save_metadata()
            print(f"Successfully installed {repo} {tag_name}")
        finally:
            # Cleanup download
            if download_path.exists():
                download_path.unlink()

    def remove(self, repo: str):
        """Removes an installed package."""
        if repo not in self.installed_packages:
            raise ValueError(f"Package {repo} is not installed.")
        
        pkg_data = self.installed_packages[repo]
        install_type = pkg_data.get("install_type")
        
        if install_type == "deb":
            apt_name = pkg_data.get("apt_name")
            if apt_name:
                self._run_sudo(["dpkg", "--remove", apt_name])
        elif install_type == "binary":
            binary_path = pkg_data.get("binary_path")
            if binary_path and os.path.exists(binary_path):
                self._run_sudo(["rm", "-f", binary_path])
                
        del self.installed_packages[repo]
        self._save_metadata()
        print(f"Successfully removed {repo}")

    def list_installed(self) -> Dict[str, Dict[str, Any]]:
        return self.installed_packages

    def _select_best_asset(self, assets: List[Dict[str, Any]]) -> Optional[Dict[str, Any]]:
        # Simple heuristic matching Go implementation logic roughly
        # Prioritize .deb, then binary, then archive
        # Filter by architecture
        system, machine = get_platform_info()
        if system != "linux":
             raise RuntimeError("This tool is strictly for Linux.")

        candidates = []
        for asset in assets:
            name = asset["name"].lower()
            # Check arch
            if machine not in name and "universal" not in name and "all" not in name:
                # If strict arch check fails, maybe skip? 
                # Go code relied on user selection often.
                # For now, let's be permissive but prioritize matches.
                pass
            
            candidates.append(asset)
            
        # Sort candidates: deb > binary > archive
        # This is a simplification; Go code had interactive selection.
        # We will implement a basic priority here.
        def score(a):
            n = a["name"].lower()
            s = 0
            if machine in n: s += 10
            if ".deb" in n: s += 5
            elif "." not in n: s += 3 # Binary usually has no ext
            elif ".tar" in n or ".zip" in n or ".gz" in n: s += 1
            return s

        candidates.sort(key=score, reverse=True)
        return candidates[0] if candidates else None

    def _download_asset(self, url: str, filename: str) -> Path:
        response = requests.get(url, stream=True)
        response.raise_for_status()
        
        fd, path = tempfile.mkstemp(prefix="get-asset-", suffix=filename)
        os.close(fd)
        file_path = Path(path)
        
        with open(file_path, "wb") as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)
        return file_path

    def _install_asset(self, file_path: Path, repo: str, release: Dict, options: Optional[Dict] = None) -> Dict[str, Any]:
        filename = file_path.name.lower()
        
        if filename.endswith(".deb"):
            return self._install_deb(file_path, options)
        elif filename.endswith((".zip", ".tar.gz", ".tgz", ".tar", ".gz")):
            return self._install_archive(file_path, repo, release, options)
        else:
            # Assume binary
            return self._install_binary(file_path, repo, options)

    def _install_deb(self, path: Path, options: Optional[Dict]) -> Dict[str, Any]:
        print("Installing .deb package with dpkg...")
        try:
            self._run_sudo(["dpkg", "-i", str(path)])
        except subprocess.CalledProcessError as e:
            # Try to fix dependencies
            print("dpkg failed, attempting to fix dependencies...")
            self._run_sudo(["apt", "-f", "install", "-y"])
            
        # Get package name
        pkg_name = self._get_deb_package_name(path)
        return {"type": "deb", "apt_name": pkg_name}

    def _get_deb_package_name(self, path: Path) -> str:
        res = subprocess.run(["dpkg-deb", "--field", str(path), "Package"], capture_output=True, text=True)
        if res.returncode == 0:
            return res.stdout.strip()
        # Fallback to filename parsing
        return path.stem.split("_")[0]

    def _install_binary(self, path: Path, repo: str, options: Optional[Dict]) -> Dict[str, Any]:
        # chmod +x
        path.chmod(0o755)
        
        binary_name = path.name
        if options and options.get("rename"):
            binary_name = options["rename"]
            
        target_path = Path("/usr/local/bin") / binary_name
        
        print(f"Installing binary to {target_path}...")
        self._run_sudo(["cp", str(path), str(target_path)])
        
        return {"type": "binary", "binary_path": str(target_path)}

    def _install_archive(self, path: Path, repo: str, release: Dict, options: Optional[Dict]) -> Dict[str, Any]:
        with tempfile.TemporaryDirectory(prefix="get-extracted-") as tmpdir:
            print(f"Extracting archive to {tmpdir}...")
            shutil.unpack_archive(path, tmpdir)
            
            # Scan for .deb or binary
            deb_path = None
            binary_path = None
            
            for root, dirs, files in os.walk(tmpdir):
                for f in files:
                    fp = Path(root) / f
                    if f.endswith(".deb"):
                        deb_path = fp
                        break
                    # Simple binary check: no extension or executable (hard to check on extraction sometimes)
                    # Check if it matches repo name loosely?
                    if not deb_path and "." not in f: # naive binary check
                        binary_path = fp
                if deb_path: break
            
            if deb_path:
                print(f"Found .deb in archive: {deb_path.name}")
                return self._install_deb(deb_path, options)
            elif binary_path:
                print(f"Found binary in archive: {binary_path.name}")
                return self._install_binary(binary_path, repo, options)
            else:
                raise ValueError("No installable .deb or binary found in archive.")

    def _run_sudo(self, cmd: List[str]):
        subprocess.run(["sudo"] + cmd, check=True)
