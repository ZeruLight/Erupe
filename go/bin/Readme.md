To bring up fresh database:
migrate.exe -database postgres://user:password@host:port/dbname?sslmode=disable -path /pathto/migrations up
To tear down database
migrate.exe -database postgres://user:password@host:port/dbname?sslmode=disable -path /pathto/migrations down


More info:
https://github.com/golang-migrate/migrate/releases/tag/v4.15.2