#!/bin/bash

###
# Runs all mitras microservices (must be previously built and installed).
#
# Expects that PostgreSQL and needed messaging DB are alredy running.
# Additionally, MQTT microservice demands that Redis is up and running.
#
###

BUILD_DIR=../build

# Kill all mitras-* stuff
function cleanup {
    pkill mitras
    pkill nats
}

###
# NATS
###
nats-server &
counter=1
until fuser 4222/tcp 1>/dev/null 2>&1;
do
    sleep 0.5
    ((counter++))
    if [ ${counter} -gt 10 ]
    then
        echo "NATS failed to start in 5 sec, exiting"
        exit 1
    fi
    echo "Waiting for NATS server"
done

###
# Users
###
MITRAS_USERS_LOG_LEVEL=info MITRAS_USERS_HTTP_PORT=9002 MITRAS_USERS_GRPC_PORT=7001 MITRAS_USERS_ADMIN_EMAIL=admin@mitras.com MITRAS_USERS_ADMIN_PASSWORD=12345678 MITRAS_USERS_ADMIN_USERNAME=admin MITRAS_EMAIL_TEMPLATE=../docker/templates/users.tmpl $BUILD_DIR/mitras-users &

###
# Clients
###
MITRAS_CLIENTS_LOG_LEVEL=info MITRAS_CLIENTS_HTTP_PORT=9000 MITRAS_CLIENTS_AUTH_GRPC_PORT=7000 MITRAS_CLIENTS_AUTH_HTTP_PORT=9002 $BUILD_DIR/mitras-clients &

###
# HTTP
###
MITRAS_HTTP_ADAPTER_LOG_LEVEL=info MITRAS_HTTP_ADAPTER_PORT=8008 MITRAS_CLIENTS_AUTH_GRPC_URL=localhost:7000 $BUILD_DIR/mitras-http &

###
# WS
###
MITRAS_WS_ADAPTER_LOG_LEVEL=info MITRAS_WS_ADAPTER_HTTP_PORT=8190 MITRAS_CLIENTS_AUTH_GRPC_URL=localhost:7000 $BUILD_DIR/mitras-ws &

###
# MQTT
###
MITRAS_MQTT_ADAPTER_LOG_LEVEL=info MITRAS_CLIENTS_AUTH_GRPC_URL=localhost:7000 $BUILD_DIR/mitras-mqtt &

###
# CoAP
###
MITRAS_COAP_ADAPTER_LOG_LEVEL=info MITRAS_COAP_ADAPTER_PORT=5683 MITRAS_CLIENTS_AUTH_GRPC_URL=localhost:7000 $BUILD_DIR/mitras-coap &

trap cleanup EXIT

while : ; do sleep 1 ; done
