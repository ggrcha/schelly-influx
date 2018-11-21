#!/bin/bash
set +e
# set +x

cd /bin

echo "Starting Influx API..."
./schelly-influx \
    --listen-ip=$LISTEN_IP \
    --listen-port=$LISTEN_PORT \
    --log-level=$LOG_LEVEL \
    --pre-post-timeout=$PRE_POST_TIMEOUT \
    --pre-backup-command="$PRE_BACKUP_COMMAND" \
    --post-backup-command="$POST_BACKUP_COMMAND" \
    --database="$DATABASE_NAME" \
    --host="$DATABASE_CONNECTION_HOST" \
    --port="$DATABASE_CONNECTION_PORT" \