#!/bin/sh
set -e

# Run dynamic auto-tuner before invoking PostgreSQL engine
if [ -f "/docker-custom/autotune.sh" ]; then
    sh /docker-custom/autotune.sh
fi

# Execute standard docker entrypoint
exec docker-entrypoint.sh "$@"
