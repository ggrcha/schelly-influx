# schelly-influx

# Usage

docker-compose .yml

```yml
version: '3.5'

services:

  db:
    image: influxdb

  schelly:
    image: flaviostutz/schelly
    ports:
      - 8080:8080
    environment:
      - LOG_LEVEL=debug
      - BACKUP_NAME=schelly-influx
      - WEBHOOK_URL=http://schelly-influx-provider:7070/backups
      - BACKUP_CRON_STRING=0 */1 * * * *
      - RETENTION_MINUTELY=5
      - WEBHOOK_GRACE_TIME=20

  schelly-influx-provider:
    image: ggrcha/schelly-influx
    build: .
    ports:
      - 7070:7070
    environment:
      - LOG_LEVEL=debug
      - BACKUP_FILE_PATH=/var/backups
      - DATABASE_NAME=schelly
      - DATABASE_CONNECTION_HOST=db
      - DATABASE_CONNECTION_PORT=8088

networks:
  default:
    name: schelly-influx-net
```

```shell
# create a new backup
curl -X POST http://localhost:7070/backups

# list existing backups
curl -X GET http://localhost:7070/backups

# get info about an specific backup
curl _X GET http://localhost:7070/backups/abc123

# remove existing backup
curl -X DELETE localhost:7070/backups/abc123

```

## REST Endpoints

As in https://github.com/flaviostutz/schelly#webhook-spec

## `influxd backup` parameters that can be set

```shell
General options:
  --file=FILENAME          output file or directory name

Options controlling the output content:
  --data-only              dump only the data, not the schema

Connection options:
  --dbname=DBNAME      database to dump (required)
  --host=HOSTNAME      database server host or socket directory (required)
  --port=PORT          database server port number

```


# Known limitations

Currently this Provider supports only synchronous backup process
