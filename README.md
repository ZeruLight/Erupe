# Erupe

## Client Compatiblity
### Platforms
- PC
- PlayStation 3
- PlayStation Vita
- Wii U (Up to Z2)
### Versions (ClientMode)
- All versions after HR compression (G10-ZZ) have been tested extensively and have great functionality.
- All versions available on Wii U (G3-Z2) have been tested and should have good functionality.
- The second oldest found version is Forward.4 (FW.4), this version has basic functionality.
- The oldest found version is Season 6.0 (S6.0), however functionality is very limited.

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

## Docker

Please see the readme in [docker/README.md](./docker/README.md). At the moment this is only really good for quick installs and checking out development not for production. 

## Schemas 

We source control the following schemas: 
- Initialisation Schemas: These initialise the application database to a clean install from a specific version.
- Update Schemas: These are update files they should be ran in order of version to get to the latest schema.
- Patch Schemas: These are for development and should be ran from the lastest available update schema or initial schema. These eventually get condensed into `Update Schemas` and then deleted when updated to a new version.
- Bundled Schemas: These are demo reference files to allow servers to be able to roll their own shops, distributions gachas and scenarios set ups. 

Note: Patch schemas are subject to change! You should only be using them if you are following along with development. 

## Resources

- [Quest and Scenario Binary Files](https://files.catbox.moe/xf0l7w.7z)
- [Mezeporta Square Discord](https://discord.gg/DnwcpXM488)
