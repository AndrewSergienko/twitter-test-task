#!/bin/bash

CERTS_DIR=/certs

echo "Creating certificates..."
rm -rf $CERTS_DIR/*

cockroach cert create-ca --certs-dir=$CERTS_DIR --ca-key=$CERTS_DIR/ca.key
cockroach cert create-node cockroachdb1 cockroachdb2 cockroachdb3 localhost 127.0.0.1 --certs-dir=$CERTS_DIR --ca-key=$CERTS_DIR/ca.key
cockroach cert create-client root --certs-dir=$CERTS_DIR --ca-key=$CERTS_DIR/ca.key

chown -R 600 $CERTS_DIR

