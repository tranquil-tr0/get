import typer
from typing import Optional
from rich.console import Console
from rich.table import Table
import sys

from .manager import PackageManager
from .utils import parse_repo_url, get_platform_info

app = typer.Typer(help="A package manager for GitHub releases (Linux Only)")
console = Console()

# Check platform on startup
system, _ = get_platform_info()
if system != "linux":
    # Allow running on other platforms for help/dev, but warn
    pass

pm = PackageManager()

@app.command()
def install(
    repo: str = typer.Argument(..., help="GitHub repository URL or user/repo"),
    release: str = typer.Option("latest", "--release", "-r", help="Specify a release version to install"),
    tag_prefix: str = typer.Option("", "--tag-prefix", "-t", help="Omit releases without the prefix"),
    rename: str = typer.Option("", "--rename", help="Rename the installed binary (binaries only)")
):
    """Install a package from GitHub."""
    if system != "linux":
        console.print("[red]Error: This tool is strictly for Linux systems.[/red]")
        raise typer.Exit(code=1)

    try:
        repo_id = parse_repo_url(repo)
        options = {"tag_prefix": tag_prefix, "rename": rename}
        pm.install(repo_id, release, options)
        console.print(f"[green]Successfully installed {repo_id}[/green]")
    except Exception as e:
        console.print(f"[red]Error installing package: {e}[/red]")
        raise typer.Exit(code=1)

@app.command("list")
def list_packages():
    """List installed packages."""
    packages = pm.list_installed()
    if not packages:
        console.print("No packages installed.")
        return

    table = Table(title="Installed Packages")
    table.add_column("Package", style="cyan")
    table.add_column("Version", style="green")
    table.add_column("Type", style="yellow")
    table.add_column("Installed At", style="magenta")

    for pkg_id, data in packages.items():
        table.add_row(
            pkg_id, 
            data.get("version", "?"), 
            data.get("install_type", "?"),
            data.get("installed_at", "?")
        )

    console.print(table)

@app.command()
def remove(repo: str = typer.Argument(..., help="GitHub repository URL or user/repo")):
    """Remove an installed package."""
    if system != "linux":
        console.print("[red]Error: This tool is strictly for Linux systems.[/red]")
        raise typer.Exit(code=1)

    try:
        repo_id = parse_repo_url(repo)
        pm.remove(repo_id)
        console.print(f"[green]Successfully removed {repo_id}[/green]")
    except Exception as e:
        console.print(f"[red]Error removing package: {e}[/red]")
        raise typer.Exit(code=1)

@app.command()
def update():
    """Check for updates."""
    console.print("[yellow]Update functionality not yet fully implemented.[/yellow]")

@app.command()
def upgrade():
    """Upgrade installed packages."""
    console.print("[yellow]Upgrade functionality not yet fully implemented.[/yellow]")

if __name__ == "__main__":
    app()
