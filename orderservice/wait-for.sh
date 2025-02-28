#!/bin/sh
# wait-for-postgres.sh

set -e

hostport="$1"
shift

host=$(echo $hostport | cut -d: -f1)
port=$(echo $hostport | cut -d: -f2)

until nc -z "$host" "$port"; do
  >&2 echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done

>&2 echo "PostgreSQL is up - executing command"
exec "$@"