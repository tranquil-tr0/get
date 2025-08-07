# Get

A tool to install and manage packages from GitHub, helping you avoid outdated or unused packages cluttering your system.

## Install

### CLI

The release is a standalone binary. Simply download, place it in your `$PATH`, and start using.
You can also download with the install script:
```sh
bash <(curl -fsSL https://raw.githubusercontent.com/tranquil-tr0/get/refs/heads/main/install.sh)
```
This will install to `/usr/local/bin`, as will installed binaries.

Run `get install tranquil-tr0/get` to start it tracking itself.

**Uninstall with:**
```sh
rm /usr/local/bin/get
```
You should also delete the json file in `~/.local/share/get/`

### GUI

The GUI is provided as a .deb file, which you can install. You can also install it with the cli by running `get install tranquil-tr0/get`

The GUI is a binary, but has dependencies:

<details><summary>Dependencies</summary>libQt6Widgets.so.6 libQt6Gui.so.6 libQt6Core.so.6 libstdc++.so.6 libm.so.6 libgcc_s.so.1 libc.so.6 libEGL.so.1 libfontconfig.so.1 libX11.so.6 libglib-2.0.so.0 libQt6DBus.so.6 libxkbcommon.so.0 libGLX.so.0 libOpenGL.so.0 libpng16.so.16 libharfbuzz.so.0 libmd4c.so.0 libfreetype.so.6 libz.so.1 libicui18n.so.76 libicuuc.so.76 libdouble-conversion.so.3 libb2.so.1 libpcre2-16.so.0 libzstd.so.1 libGLdispatch.so.0 libexpat.so.1 libxcb.so.1 libatomic.so.1 libpcre2-8.so.0 libdbus-1.so.3 libgraphite2.so.3 libbz2.so.1.0 libbrotlidec.so.1 libicudata.so.76 libgomp.so.1 libXau.so.6 libXdmcp.so.6 libsystemd.so.0 libbrotlicommon.so.1 libcap.so.2</details>

You can run the gui after installing it and its dependencies
The gui has a few less features than the cli, but is generally usable

## How do I use this?
You can learn more about each command by running `--help` - for example:
`get --help`
`get install --help`
`get update --help`

## Contributing
Issues are welcome!
PRs are welcome!

build commands:
`go build -o get ./cmd/cli/main.go`
`go build -o get-gui -ldflags "-s -w" ./cmd/gui/main.go`
`sudo dpkg-deb --build --root-owner-group get-gui-deb`

![image](https://github.com/user-attachments/assets/87672626-1f60-4ec5-9358-b539b8a5d79c)

---
built with go, conda, and miqt
