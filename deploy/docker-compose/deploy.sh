#!/bin/bash

DIR=$( cd $(dirname $0) ; pwd -P )
ROOT_DIR=$DIR/../../

echo "Running services..."
GUNGNIR_VERSION=${GUNGNIR_VERSION:-0.12.3} \
SVALINN_VERSION=${SVALINN_VERSION:-0.14.0} \
docker-compose -f $ROOT_DIR/deploy/docker-compose/docker-compose.yml up -d $@
sleep 5

docker exec -it yb-tserver-n1 /home/yugabyte/bin/cqlsh yb-tserver-n1 -f /create_db.cql