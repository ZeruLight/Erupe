# Erupe

Nothing here yet :)

Based on the TW version. Requires a local mirror of the original launcher site to be placed in `./www/g6_launcher` until I can RE the launcher and figure out which JS callbacks it requires.


## Installation
Clone the repo
Install PostgreSQL, launch psql shell, `CREATE DATABASE erupe;`.

Setup db with golang-migrate:
`go get -tags 'postgres' -u github.com/golang-migrate/migrate/cmd/migrate`
`set POSTGRESQL_URL=postgres://postgres:password@localhost:5432/erupe?sslmode=disable`
`migrate -database %POSTGRESQL_URL% -path migrations up`
