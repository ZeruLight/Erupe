# Erupe
## WARNING 
This project is in it's infancy and currently doesn't do anything worth noting. Additionally, it has no documentation, no support, and cannot be used without binary resources that are not in the repo. 

# General info
Based on the TW version. Requires a local mirror of the original launcher site to be placed in `./www/g6_launcher` until I can RE the launcher and figure out which JS callbacks it requires.

## Installation
Clone the repo
Install PostgreSQL, launch psql shell, `CREATE DATABASE erupe;`.

Setup db with golang-migrate:
`go get -tags 'postgres' -u github.com/golang-migrate/migrate/cmd/migrate`
`set POSTGRESQL_URL=postgres://postgres:password@localhost:5432/erupe?sslmode=disable`
`migrate -database %POSTGRESQL_URL% -path migrations up`

Open psql shell and manually insert an account into the users table.

Run `test.py` with python 3 to generate an entrance server response binary because the code isn't ported to Go yet.

Place a copy of the original TW launcher html/js/css in `./www/g6_launcher/`, and a copy of the serverlist at `./www/server/serverlist.xml`.

Manually extract the binary response from a pcap, strip the header, and decrypt the ~50 packets that are used in `./channelserver/session.go`, and place them in `./bin_resp/{OPCODE}_resp.bin`.


# Use
Add to hosts:
```
127.0.0.1 mhfg.capcom.com.tw
127.0.0.1 mhf-n.capcom.com.tw
```
