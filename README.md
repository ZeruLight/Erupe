# Erupe

## Client Compatiblity
### Platforms
- PC
- PlayStation 3
- PlayStation Vita
- Wii U (Up to Z2)
### Versions (ClientMode)
All versions after HR compression (G10-ZZ) have been tested extensively and have great functionality.
All versions available on Wii U (G3-Z2) have been tested and should have good functionality.
The second oldest found version is Forward.4 (FW.4), this version has basic functionality.
The oldest found version is Season 6.0 (S6.0), however functionality is very limited.

If you have an **installed** copy of Monster Hunter Frontier on an old hard drive, **please** get in contact so we can archive it!

## Setup

If you are only looking to install Erupe, please use [a pre-compiled binary](https://github.com/ZeruLight/Erupe/releases/latest).

If you want to modify or compile Erupe yourself, please read on.

## Requirements

- [Go](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)

## Installation

1. Bring up a fresh database by using the [backup file attached with the latest release](https://github.com/ZeruLight/Erupe/releases/latest/download/SCHEMA.sql).
2. Run each script under [patch-schema](./patch-schema) as they introduce newer schema.
3. Edit [config.json](./config.json) such that the database password matches your PostgreSQL setup.
4. Run `go build` or `go run .` to compile Erupe.

## Resources

- [Quest and Scenario Binary Files](https://files.catbox.moe/xf0l7w.7z)
- [Mezeporta Square Discord](https://discord.gg/DnwcpXM488)
