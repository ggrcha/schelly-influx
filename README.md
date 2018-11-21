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
      - 7070:7070
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
curl -X GET http://localhost:7070/backups/abc123

# remove existing backup
curl -X DELETE http://localhost:7070/backups/abc123

```

## REST Endpoints

As in https://github.com/flaviostutz/schelly#webhook-spec

## `influxd backup` parameters that can be set

```shell
General options:
  --retention=retention retention policy for the backup. If not specified, the default is to use all retention policies.
  --shard=shard         shard ID of the shard to be backed up
  --start=start         include all points starting with the specified timestamp (RFC3339 format)
	--end=end             exclude all results after the specified timestamp (RFC3339 format)
	--since=since         perform an incremental backup after the specified timestamp RFC3339 format

Connection options:
  --database=DBNAME     database to dump (required)
  --host=HOSTNAME       database server host or socket directory (required)
  --port=PORT           database server port number

```


# Known limitations

Currently this Provider supports only synchronous backup process
