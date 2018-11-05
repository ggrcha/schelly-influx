# schelly-influx

# Usage

docker-compose .yml

```yml
version: '3.5'

services:

  db:
    image: postgres
    environment:
      - POSTGRES_DB=schelly
      - POSTGRES_PASSWORD=postgres

  schelly:
    image: flaviostutz/schelly
    ports:
      - 8080:8080
    environment:
      - LOG_LEVEL=debug
      - BACKUP_NAME=schelly-pgdump
      - WEBHOOK_URL=http://schelly-postgres-provider:7070/backups
      - BACKUP_CRON_STRING=0 */1 * * * *
      - RETENTION_MINUTELY=5
      - WEBHOOK_GRACE_TIME=20

  schelly-postgres-provider:
    image: tiagostutz/schelly-postgres
    build: .
    ports:
      - 7070:7070
    environment:
      - LOG_LEVEL=debug
      - BACKUP_FILE_PATH=/var/backups
      - DATABASE_NAME=schelly
      - DATABASE_CONNECTION_HOST=db
      - DATABASE_CONNECTION_PORT=5432
      - DATABASE_AUTH_USERNAME=postgres
      - DATABASE_AUTH_PASSWORD=postgres

networks:
  default:
    name: schelly-postgres-net
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

## `pg_dump` parameters that can be set

```shell
General options:
  --file=FILENAME          output file or directory name

Options controlling the output content:
  --data-only              dump only the data, not the schema
  --encoding=ENCODING      dump the data in encoding ENCODING
  --schema-only            dump only the schema, no data  

Connection options:
  --dbname=DBNAME      database to dump (required)
  --host=HOSTNAME      database server host or socket directory (required)
  --port=PORT          database server port number
  --username=NAME      connect as specified database user (defaults to "postgres")

Schelly postgres provider custom options:
  --password           password to be placed on ~/.pgpass (required)
```

`pg_dump` parameters that currently can't be set and the values that are used:

```
General options:
  --format=c|d|t|p         output file format (custom, directory, tar,
                               plain text (default))                                 -> value used: p
  --jobs=NUM               use this many parallel jobs to dump                      -> value used: 1
  --verbose                verbose mode                                             -> value used: --verbose  
  --compress=0-9           compression level for compressed formats                 -> value used: 9
  --column-inserts             dump data as INSERT commands with column names       -> value used: --column-inserts
  --inserts                    dump data as INSERT commands, rather than COPY       -> value used: --inserts  
  --quote-all-identifiers      quote all identifiers, even if not key words         -> value used: --quote-all-identifiers

Options controlling the output content:
  --clean                  clean (drop) database objects before recreating     -> value used: --clean  
  --create                 include commands to create database in dump         -> value used: --create  
```

## `pg_dump` parameters that currently can't be set
```
  --schema=SCHEMA          dump the named schema(s) only  
  --exclude-schema=SCHEMA  do NOT dump the named schema(s)  
  --table=TABLE            dump the named table(s) only  
  --exclude-table=TABLE    do NOT dump the named table(s)  
  --exclude-table-data=TABLE   do NOT dump data for the named table(s)
  --no-password        never prompt for password
  
```

# Known limitations

Currently this Provider supports only synchronous backup process
