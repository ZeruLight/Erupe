# Erupe Community Edition

## Setup

If you are only looking to install Erupe, please use [a pre-compiled binary](https://github.com/ZeruLight/Erupe/releases/latest).

If you want to modify or compile Erupe yourself, please read on.

### Requirements

- [Go](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)

### Installation

1. Bring up a fresh database by using the [backup file attached with the latest release](https://github.com/ZeruLight/Erupe/releases/latest/download/SCHEMA.sql).
2. Run each script under [patch-schema](./patch-schema) as they introduce newer schema.
3. Edit [config.json](./config.json) such that the database password matches your PostgreSQL setup.
4. Run `go build` or `go run .` to compile Erupe.

### Note

- You will need to acquire and install the client files and quest binaries separately.

## Resources

- [Quest and Scenario Binary Files](https://files.catbox.moe/xf0l7w.7z)
- [PewPewDojo Discord](https://discord.gg/CFnzbhQ)
- [Community FAQ Pastebin](https://pastebin.com/QqAwZSTC)ter
