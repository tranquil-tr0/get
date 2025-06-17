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

---
built with go and conda