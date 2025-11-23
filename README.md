# Get

A tool to install and manage packages from GitHub releases, helping you avoid outdated or unused packages cluttering your system.
Rewritten in Python with a Kirigami GUI.

> [!IMPORTANT]
> **Linux Only**: This tool is strictly for Linux systems. It manages `.deb` packages via `dpkg` and binaries in `/usr/local/bin`.

## Install

### Prerequisites
- Linux
- Python 3.10+
- `pip` and `venv`
- `sudo` access (for installing packages)

### Installation
Clone the repository and install:

```sh
git clone https://github.com/tranquil-tr0/get.git
cd get

# Create and activate virtual environment
python -m venv .venv
source .venv/bin/activate

# Install the package
pip install .
```

## Usage

### CLI

Run the CLI using `get` (if installed in PATH):

```sh
# Install a package (requires sudo)
get install tranquil-tr0/get

# List installed packages
get list

# Remove a package (requires sudo)
get remove tranquil-tr0/get
```

### GUI

The GUI is built with KDE Kirigami. To run it:

```sh
get-gui
```

## Contributing

Issues and PRs are welcome!

### Development
1. Clone the repository
2. Create a virtual environment: `python -m venv .venv`
3. Activate it: `source .venv/bin/activate`
4. Install in development mode: `pip install -e .`
5. Run the CLI: `get --help`
6. Run the GUI: `get-gui`

