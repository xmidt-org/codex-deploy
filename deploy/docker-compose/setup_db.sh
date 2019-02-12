#!/bin/bash

HOSTPARAMS="--host db --insecure"
SQL="/cockroach/cockroach.sh sql $HOSTPARAMS"

$SQL -e "CREATE USER IF NOT EXISTS roachadmin;"
$SQL -e "CREATE DATABASE IF NOT EXISTS devices;"
$SQL -e "GRANT ALL ON DATABASE devices TO roachadmin;"