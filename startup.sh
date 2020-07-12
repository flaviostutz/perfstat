#!/bin/bash
set -e
# set -x

echo "Starting Restic API..."
schelly-restic \
    --listen-ip=$LISTEN_IP \
    --listen-port=$LISTEN_PORT \
    --log-level=$LOG_LEVEL \
    --repo-dir=$TARGET_DATA_PATH \
    --source-path=$SOURCE_DATA_PATH \
    --pre-post-timeout=$PRE_POST_TIMEOUT \
    --pre-backup-command="$PRE_BACKUP_COMMAND" \
    --post-backup-command="$POST_BACKUP_COMMAND"

