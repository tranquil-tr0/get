# Get

A CLI tool to install and manage packages from GitHub, helping you avoid outdated or unused packages cluttering your system.

## Install

The release is a standalone binary. Simply download, place it in your `$PATH`, and start using.
You can also download with the install script:
```sh
bash <(curl -fsSL https://raw.githubusercontent.com/tranquil-tr0/get/refs/heads/main/install.sh)
```
This will install to /usr/local/bin and can be uninstalled with:
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

---
built with go and conda