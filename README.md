# Erupe
## WARNING 
This project is in its infancy and currently doesn't do anything worth noting. Additionally, it has no documentation, no support, and cannot be used without binary resources that are not in the repo. 

# General info
Originally based on the TW version, but (slowly) transitioning to JP.

# Installation
1. Clone the repo with `git clone https://github.com/Andoryuuta/Erupe.git`
2. Install PostgreSQL
3. Launch psql shell, `CREATE DATABASE erupe;`.

4. Setup database with golang-migrate:
```
> go get -tags 'postgres' -u github.com/golang-migrate/migrate/cmd/migrate

> set POSTGRESQL_URL=postgres://postgres:password@localhost:5432/erupe?sslmode=disable

> migrate -database %POSTGRESQL_URL% -path migrations up
```

5. Open psql shell and manually insert an account into the users table.

6. Run `test.py` with python 3 to generate an entrance server response binary because the code isn't ported to Go yet. (**This requires a binary response that is not included in the repo**)

7. Manually extract the binary response from a pcap, strip the header, and decrypt the ~50 packets that are used in `./channelserver/session.go`, and place them in `./bin_resp/{OPCODE}_resp.bin`. (**These are not included in the repo**)




## Launcher
Erupe now ships with a rudimentary custom launcher, so you don't need to obtain the original TW/JP files to simply get ingame. However, it does still support using the original files if you choose to. To set this up, place a copy of the original launcher html/js/css in `./www/tw/`, and `/www/jp/` for the TW and JP files respectively.

Then, modify the the `/launcher/js/launcher.js` file as such:
* Find the call to `startUpdateProcess();` in a case statement and replace it with `finishUpdateProcess();`. (This disables the file check and updating)
* (JP ONLY): replace all uses of "https://" with "http://" in the file.

Finally, edit `main.go` and change:

`go serveLauncherHTML(":80", false)`

to:

`go serveLauncherHTML(":80", true)`

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
