#!/bin/bash

docker pull postgres:12
if [ ! "$(docker ps -q -f name=pgsql1)" ]; then
    if [ "$(docker ps -aq -f status=exited -f name=pgsql1)" ]; then
        docker rm pgsql1
    fi
    docker run --name=pgsql1 -p 5432:5432 -v "/opt/databases/postgres":/var/lib/postgresql/data -e POSTGRES_PASSWORD=1110 -e POSTGRES_DB=test -d postgres:12
    # ss -tulpn | grep 5432
fi
