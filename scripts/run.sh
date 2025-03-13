###
# Runs all Athena microservices (must be previously built and installed).
#
# Expects that PostgreSQL and needed messaging DB are alredy running.
# Additionally, MQTT microservice demands that Redis is up and running.
#
###

BUILD_DIR=../build

# Kill all athena-* stuff
function cleanup {
    pkill athena
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
ATHENA_USERS_LOG_LEVEL=info ATHENA_USERS_HTTP_PORT=9002 ATHENA_USERS_GRPC_PORT=7001 ATHENA_USERS_ADMIN_EMAIL=admin@athena.com ATHENA_USERS_ADMIN_PASSWORD=12345678 ATHENA_USERS_ADMIN_USERNAME=admin ATHENA_EMAIL_TEMPLATE=../docker/templates/users.tmpl $BUILD_DIR/athena-users &

###
# Clients
###
ATHENA_CLIENTS_LOG_LEVEL=info ATHENA_CLIENTS_HTTP_PORT=9000 ATHENA_CLIENTS_GRPC_PORT=7000 ATHENA_CLIENTS_AUTH_HTTP_PORT=9002 $BUILD_DIR/athena-clients &

###
# HTTP
###
ATHENA_HTTP_ADAPTER_LOG_LEVEL=info ATHENA_HTTP_ADAPTER_PORT=8008 ATHENA_CLIENTS_GRPC_URL=localhost:7000 $BUILD_DIR/athena-http &

###
# WS
###
ATHENA_WS_ADAPTER_LOG_LEVEL=info ATHENA_WS_ADAPTER_HTTP_PORT=8190 ATHENA_CLIENTS_GRPC_URL=localhost:7000 $BUILD_DIR/athena-ws &

###
# MQTT
###
ATHENA_MQTT_ADAPTER_LOG_LEVEL=info ATHENA_CLIENTS_GRPC_URL=localhost:7000 $BUILD_DIR/athena-mqtt &

###
# CoAP
###
ATHENA_COAP_ADAPTER_LOG_LEVEL=info ATHENA_COAP_ADAPTER_PORT=5683 ATHENA_CLIENTS_GRPC_URL=localhost:7000 $BUILD_DIR/athena-coap &

trap cleanup EXIT

while : ; do sleep 1 ; done