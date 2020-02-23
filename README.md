# Erupe
## WARNING 
This project is in its infancy and currently doesn't do anything worth noting. Additionally, it has no documentation and no support.

This project is soley developed in my spare time for the educational experience of making a server emulator, which I haven't done before. Expectations regarding functionally and code quality should be set accordingly.

# General info
Currently allows a JP MHF client (with GameGuard removed) to:
* Login and register an account (registration is automatic if account doesn't exist)
* Create a character
* Get ingame to the main city
* See other players walk around (Names and character data are not correct, placeholders are in use)
* Use (local) chat.

# Installation
1. Clone the repo with `git clone https://github.com/Andoryuuta/Erupe.git`
2. Install PostgreSQL
3. Launch psql shell, `CREATE DATABASE erupe;`.
4. Setup database with golang-migrate:

    Windows:
    ```
    > go get -tags 'postgres' -u github.com/golang-migrate/migrate/cmd/migrate

    > set POSTGRESQL_URL=postgres://postgres:password@localhost:5432/erupe?sslmode=disable

    > cd erupe

    > migrate -database %POSTGRESQL_URL% -path migrations up
    ```

    Linux:
    ```
    > go get -tags 'postgres' -u github.com/golang-migrate/migrate/cmd/migrate

    > export POSTGRESQL_URL=postgres://postgres:password@localhost:5432/erupe?sslmode=disable

    > cd erupe

    > migrate -database $POSTGRESQL_URL -path migrations up
    ```

    (Replacing `postgres:password` with your postgres username and password)

5. Edit the config.json
    Namely:
    * Update the database username and password
    * Update the `host_ip` and `ip` fields (there are multiple) to your external IP if you are hosting for multiple clients.


## Launcher
Erupe ships with a rudimentary custom launcher, so you don't need to obtain the original TW/JP files to simply get ingame. However, it does still support using the original files if you choose to. To set this up, place a copy of the original launcher html/js/css in `./www/tw/`, and `/www/jp/` for the TW and JP files respectively.

Then, modify the the `/launcher/js/launcher.js` file as such:
* Find the call to `startUpdateProcess();` in a case statement and replace it with `finishUpdateProcess();`. (This disables the file check and updating)
* (JP ONLY): replace all uses of "https://" with "http://" in the file.

Finally, edit the config.json and set `UseOriginalLauncherFiles` to `true` under the launcher settings.

# Usage
### Note: If you are switching to/from the custom launcher html, you will have to clear your IE cache @ `C:\Users\<user>\AppData\Local\Microsoft\Windows\INetCache`.

## Server
```
cd Erupe
go run .
```

## Client
Add to hosts:
```
127.0.0.1 mhfg.capcom.com.tw
127.0.0.1 mhf-n.capcom.com.tw
127.0.0.1 cog-members.mhf-z.jp
127.0.0.1 www.capcom-onlinegames.jp
127.0.0.1 srv-mhf.capcom-networks.jp
```

Run mhf.exe normally (with locale emulator or appropriate timezone).
