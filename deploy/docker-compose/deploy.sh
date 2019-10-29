#!/bin/bash

DIR=$( cd $(dirname $0) ; pwd -P )
ROOT_DIR=$DIR/../../

echo "Building Fenrir Locally..."
git clone https://github.com/xmidt-org/fenrir.git /tmp/fenrir
pushd /tmp/fenrir
docker build -f ./deploy/Dockerfile -t fenrir:local .
docker build -t simulator:local $ROOT_DIR/simulator
popd

echo "Running services..."
GUNGNIR_VERSION=${GUNGNIR_VERSION:-0.9.2} \
SVALINN_VERSION=${SVALINN_VERSION:-0.11.2} \
FENRIR_VERSION=${FENRIR_VERSION:-local} \
docker-compose -f $ROOT_DIR/deploy/docker-compose/docker-compose.yml up -d $@