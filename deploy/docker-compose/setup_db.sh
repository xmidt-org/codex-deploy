#!/bin/bash

HOSTPARAMS="--host db --insecure"
SQL="/cockroach/cockroach.sh sql $HOSTPARAMS"

$SQL -e "CREATE USER IF NOT EXISTS roachadmin;"
$SQL -e "CREATE DATABASE IF NOT EXISTS devices;"
$SQL -e "GRANT ALL ON DATABASE devices TO roachadmin;"
$SQL -e "CREATE TABLE devices.events (type INT NOT NULL, device_id STRING NOT NULL, birth_date BIGINT NOT NULL, death_date BIGINT NOT NULL, data BYTES NOT NULL, nonce BYTES NOT NULL, alg STRING NOT NULL, kid STRING NOT NULL, INDEX idx_events_id_birth_date (device_id, birth_date DESC), INDEX idx_events_death_date (death_date DESC)) "
$SQL -e "CREATE TABLE devices.blacklist (id STRING PRIMARY KEY, reason STRING NOT NULL);"