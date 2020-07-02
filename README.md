![gopher](gopher.png)

# Gophercraft

[![](https://godoc.org/github.com/superp00t/gophercraft?status.svg)](github.com/superp00t/gophercraft)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Chat on discord](https://img.shields.io/discord/556039662997733391.svg)](https://discord.gg/xPtuEjt)

The Gophercraft project provides 100% Go libraries and programs for research and experimentation with MMORPG software.

Gophercraft aims to provide a common API for interfacing with multiple protocols, instead of trying support every protocol version with its own Git branch. At the moment, protocol 5875 is the most supported but we are working hard to add better support for other versions.

In addition to general purpose packages, Gophercraft provides a multi-server core styled after MaNGOS.

Some caveats: Gophercraft is currently in development and **extremely** unstable: expect bugs, a general lack of features and a frequently changing API.

**⚠️ WARNING: Gophercraft is currently prone to all sorts of game-ruining exploits, and requires additional hardening before you use it for your own game.**

## What works so far in Gophercraft Core:

- Registration ✓
- Authentication ✓
- Server selection ✓
- Creating characters ✓
- Moving around in the world (5875 client) ✓
- Authentication and realm list server ✓
- Authentication protocol client ✓
- Game protocol client and server (partially)
- HTTP JSON API for facilitating registration ✓
- Support for Windows, Linux and Mac OS X ✓
- Formatting/conversion tools written in pure Go ✓
- Integrated mod manager ✓

## Server setup/installation

```bash
# Install package
sudo apt-get install git golang mariadb-server

# Create default databases
MYPWD="my password here" mysql -u root -p$MYPWD<<EOL
CREATE DATABASE gcraft_auth;
CREATE DATABASE gcraft_world_1;
EOL
```

If you want to operate multiple world servers, you must create a new database for each.

```bash
go get -u -v github.com/superp00t/gophercraft/cmd/gcraft_wizard
go get -u -v github.com/superp00t/gophercraft/cmd/gcraft_core_auth
go get -u -v github.com/superp00t/gophercraft/cmd/gcraft_core_world

# Autogenerate configuration files
gcraft_wizard

# Edit your configurations in ~/.local/share/gcraft_auth/config.yml
#                      and in ~/.local/share/gcraft_world_1/config.yml
# launch authserver (do this in background)
gcraft_core_auth

# launch worldserver
gcraft_core_world
```

To register, point your browser to http://localhost:8086 and fill out the registration form.
