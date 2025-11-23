# Get

A tool to install and manage packages from GitHub releases, helping you avoid outdated or unused packages cluttering your system.
Rewritten in Python with a Kirigami GUI.

> [!IMPORTANT]
> **Linux Only**: This tool is strictly for Linux systems. It manages `.deb` packages via `dpkg` and binaries in `/usr/local/bin`.

## Install

### Prerequisites
- Linux
- Python 3.10+
- `uv` (recommended) or `pip`
- `sudo` access (for installing packages)

### Installation
Clone the repository and install with `uv`:

```sh
git clone https://github.com/tranquil-tr0/get.git
cd get
uv sync
```

Or install directly with `pip`:

```sh
pip install .
```

## Usage

### CLI

Run the CLI using `get` (if installed in PATH) or via `uv run`:

```sh
# Install a package (requires sudo)
uv run get install tranquil-tr0/get

# List installed packages
uv run get list

# Remove a package (requires sudo)
uv run get remove tranquil-tr0/get
```

### GUI

The GUI is built with KDE Kirigami. To run it:

```sh
uv run get-gui
```

## Contributing

Issues and PRs are welcome!

### Development
1. Install `uv`.
2. Run `uv sync` to install dependencies.
3. Run `uv run get --help` to test the CLI.
