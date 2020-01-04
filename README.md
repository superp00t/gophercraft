![gopher](gopher.png)

# Gophercraft

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
- Integrated mod manager (partially, in the form of datapacks)

# What needs to be created
- Backpacks (automatic backup datapacks)
- Cheating, fuzzing and penetration testing tools, to audit Gophercraft's resilience to malicious clients (code review by people knowledgeable about exploits would be appreciated)
- Scripting/AI system
- Full support for All-GM Roleplay gamemode
- Item/NPC forge
- Helper AddOn + bidirectional RPC system through AddOn channel
- Implement geometry checks so players can't teleport out of bounds
- Rich web application utilizing the Gophercraft JSON API, browsing players, stats, items and guilds

## Server setup/installation

Gophercraft uses [xorm](https://xorm.io/) for storing data.

MaNGOS-based cores require the installation of an entire SQL database to use their content, whereas Gophercraft Core uses [datapacks.](datapack).

Datapacks are ZIP archives (alternatively uncompressed folders), which containerize both tabular data (CSV) and scripting (Lua). Upon the launch of the worldserver, these datapacks are combined and dynamically loaded into your SQL database, where they can be quickly queried. This makes them ideal for making modifications to the game.

Currently, Gophercraft permits only the use of MySQL/MariaDB as a connection backend. [SQLite3](https://github.com/superp00t/gophercraft/blob/master/gcore/dbsupport/sqlite3.go) is currently broken due to the limitations of the C implementation, although other backends supported by xorm may be theoretically usable.

If you'd like to inquire about adding support for a new database, open an issue or a pull request!

```bash
# install packages
sudo apt-get install git golang mysql-server

# create default databases
cat >/tmp/gcraft_gen.sql <<EOL
CREATE DATABASE gcraft_auth;
CREATE DATABASE gcraft_world_1;
EOL

cat /tmp/gcraft_gen.sql | mysql -u root -p
```

Only one auth database is used. If you want to operate multiple world servers, you must create a new database for each. Coordination between servers is accomplished through GRPC endpoints.

```bash
go get -u -v github.com/superp00t/gophercraft/cmd/gcraft_wizard
go get -u -v github.com/superp00t/gophercraft/cmd/gcraft_core_auth
go get -u -v github.com/superp00t/gophercraft/cmd/gcraft_core_world

# generate config and datapack based on game
gcraft_wizard -w /path/to/game/

# Edit your configurations in ~/.local/share/gcraft_auth/config.yml
#                      and in ~/.local/share/gcraft_world_1/config.yml
# launch authserver (do this in background)
gcraft_core_auth

# launch worldserver
gcraft_core_world
```

To register, point your browser to http://localhost:8086 and fill out the registration form.

It's important to do this before exposing your server, as the first user to register is granted administrator privileges.
