# Docker for erupe

## Building the container
Run the following from the route of the soruce folder. In this example we give it the tag of dev to seperate it from any other container verions. 
```bash
docker build . -t erupe:dev
```
## Running the container in isolation
This is just running the container. You can do volume mounts into the container for the `config.json` to tell it to communicate to a database. You will need to do this also for other folders such as `bin` and `savedata`
```bash
docker run erupe:dev
```

## Docker compose
Docker compose allows you to run multiple containers at once. The docker compose in this folder has 3 things set up.
- postgres
- pg admin (Admin interface to make db changes)
- erupe

Before we get started you should make sure the database info matches whats in the docker compose file for the environment variables `POSTGRES_PASSWORD`,`POSTGRES_USER` and `POSTGRES_DB`. You can set the host to be the service name `db`.

Here is a example of what you would put in the config.json if you was to leave the defaults. It is strongly recommended to change the password. 
```txt
"Database": {
    "Host": "db",
    "Port": 5432,
    "User": "postgres",
    "Password": "password",
    "Database": "erupe"
  },
```

### Running up the database for the first time
First we need to set up the database. This requires the schema and the patch schemas to be applied. This can be done by runnnig up both the db and pgadmin.

1. Pull the remote images and build a container image for erupe
```bash
docker-compose pull 
docker-compose build
```
2. Run up pgadmin and login using the username and password provided in `PGADMIN_DEFAULT_EMAIL` and `PGADMIN_DEFAULT_PASSWORD` note you will need to set up a new connection to the database internally. You will use the same host, database, username and password as above. 
```bash
docker-compose run db pgadmin -d
```
3. Use pgadmin to restore the schema using the restore functionaltiy and they query tool for the patch-schemas.

4. Now run up the server you should see the server start correctly now. 
```bash
docker-compose run server -d
```

## Turning off the server safely
```bash
docker-compose stop
```
## Turning on the server again 
This boots the db pgadmin and the server in a detached state
```bash
docker-compose up -d
```
if you want all the logs and you want it to be in an attached state 
```bash
docker-compose up
```
