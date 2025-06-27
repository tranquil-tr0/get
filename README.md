# Get

A CLI tool to install and manage packages from GitHub, helping you avoid outdated or unused packages cluttering your system.

## Install

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

## How do I use this?
You can learn more about each command by running `--help` - for example:
`get --help`
`get install --help`
`get update --help`

## Contributing
feel free to file issues and PRs

![image](https://github.com/user-attachments/assets/87672626-1f60-4ec5-9358-b539b8a5d79c)

---
built with go and conda
