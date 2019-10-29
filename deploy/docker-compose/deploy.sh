#!/bin/bash

DIR=$( cd $(dirname $0) ; pwd -P )
ROOT_DIR=$DIR/../../

echo "Running services..."
GUNGNIR_VERSION=${GUNGNIR_VERSION:-0.10.1-rc.1} \
SVALINN_VERSION=${SVALINN_VERSION:-0.12.0} \
docker-compose -f $ROOT_DIR/deploy/docker-compose/docker-compose.yml up -d $@
