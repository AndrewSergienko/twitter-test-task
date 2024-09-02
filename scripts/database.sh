#!/bin/bash

CERTS_DIR=/certs

echo "Creating certificates..."
if [[ -f $CERTS_DIR/ca.crt && -f $CERTS_DIR/ca.key && -f $CERTS_DIR/client.root.crt && -f $CERTS_DIR/client.root.key && -f $CERTS_DIR/node.crt && -f $CERTS_DIR/node.key ]]; then
    echo "Certificates already exist"
else
    echo "Certificates do not exist, creating..."
    rm -rf $CERTS_DIR/*

    cockroach cert create-ca --certs-dir=$CERTS_DIR --ca-key=$CERTS_DIR/ca.key
    cockroach cert create-node cockroachdb1 cockroachdb2 cockroachdb3 localhost 127.0.0.1 --certs-dir=$CERTS_DIR --ca-key=$CERTS_DIR/ca.key
    cockroach cert create-client root --certs-dir=$CERTS_DIR --ca-key=$CERTS_DIR/ca.key

    chown -R 600 $CERTS_DIR
fi
