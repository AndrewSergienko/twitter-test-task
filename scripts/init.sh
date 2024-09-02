#!/bin/bash

CERTS_DIR=/cockroach/certs

if [ -f ./conf/.env ]; then
    export $(grep -v '^#' ./conf/.env | xargs)
fi

docker compose -f ./docker-compose.yml --profile setup down
docker compose -f ./docker-compose.yml --profile main down
docker compose -f ./docker-compose.yml --profile setup up -d

echo "Waiting for the cluster to start..."
sleep 5

if ! docker ps --filter "name=cockroachdb1" --filter "status=running" | grep -q cockroachdb1; then
  echo "Container cockroachdb1 is not running. Exiting."
  exit 1
fi

echo "Initializing cluster..."

docker exec cockroachdb1 cockroach init --certs-dir=$CERTS_DIR --host=localhost:26257

echo "Creating database user..."
docker exec cockroachdb1 cockroach sql --certs-dir=$CERTS_DIR --host=cockroachdb1:26257 --execute "CREATE USER $COCKROACH_USER WITH PASSWORD '$COCKROACH_PASSWORD';"
docker exec cockroachdb1 cockroach cert create-client $COCKROACH_USER --certs-dir=$CERTS_DIR --ca-key=$CERTS_DIR/ca.key

echo "Creating database..."
docker exec cockroachdb1 cockroach sql --certs-dir=$CERTS_DIR --host=cockroachdb1:26257 --execute "CREATE DATABASE $COCKROACH_DATABASE;"
docker exec cockroachdb1 cockroach sql --certs-dir=$CERTS_DIR --host=cockroachdb1:26257 --execute "GRANT ALL ON DATABASE $COCKROACH_DATABASE TO $COCKROACH_USER;"

echo "Database setup complete"
docker compose -f ./docker-compose.yml --profile main up -d

if [ "$BOT_ENABLE" = "true" ]; then
  echo "BOT_ENABLE встановлено на true. Запускаю бота..."

  docker compose -f ./docker-compose.yml --profile bot up -d
fi
