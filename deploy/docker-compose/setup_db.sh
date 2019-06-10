#!/bin/bash

HOSTPARAMS="--host db --insecure"
SQL="/cockroach/cockroach.sh sql $HOSTPARAMS"

$SQL -e "CREATE USER IF NOT EXISTS roachadmin;"
$SQL -e "CREATE DATABASE IF NOT EXISTS devices;"
$SQL -e "GRANT ALL ON DATABASE devices TO roachadmin;"
$SQL -e "CREATE TABLE devices.events (device_id STRING NOT NULL, type INT8 NOT NULL, birth_date INT8 NOT NULL, record_id INT8 NOT NULL DEFAULT unique_rowid(), death_date INT8 NOT NULL, data BYTES NOT NULL, nonce BYTES NOT NULL, alg STRING NOT NULL, kid STRING NOT NULL, shard INT2 AS ((fnv32(device_id) % 8)::INT2) STORED CHECK (shard IN (0, 1, 2, 3, 4, 5, 6, 7)), PRIMARY KEY (device_id, type, birth_date DESC, record_id), INDEX idx_events_death_date (shard, death_date DESC, record_id));"
$SQL -e "CREATE TABLE devices.blacklist (id STRING PRIMARY KEY, reason STRING NOT NULL);"