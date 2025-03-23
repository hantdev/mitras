if [ -z "$MITRAS_MQTT_CLUSTER" ]
then
      envsubst '${MITRAS_MQTT_ADAPTER_MQTT_PORT}' < /etc/nginx/snippets/mqtt-upstream-single.conf > /etc/nginx/snippets/mqtt-upstream.conf
      envsubst '${MITRAS_MQTT_ADAPTER_WS_PORT}' < /etc/nginx/snippets/mqtt-ws-upstream-single.conf > /etc/nginx/snippets/mqtt-ws-upstream.conf
else
      envsubst '${MITRAS_MQTT_ADAPTER_MQTT_PORT}' < /etc/nginx/snippets/mqtt-upstream-cluster.conf > /etc/nginx/snippets/mqtt-upstream.conf
      envsubst '${MITRAS_MQTT_ADAPTER_WS_PORT}' < /etc/nginx/snippets/mqtt-ws-upstream-cluster.conf > /etc/nginx/snippets/mqtt-ws-upstream.conf
fi

envsubst '
    ${MITRAS_NGINX_SERVER_NAME}
    ${MITRAS_AUTH_HTTP_PORT}
    ${MITRAS_DOMAINS_HTTP_PORT}
    ${MITRAS_GROUPS_HTTP_PORT}
    ${MITRAS_USERS_HTTP_PORT}
    ${MITRAS_CLIENTS_HTTP_PORT}
    ${MITRAS_CLIENTS_AUTH_HTTP_PORT}
    ${MITRAS_CHANNELS_HTTP_PORT}
    ${MITRAS_HTTP_ADAPTER_PORT}
    ${MITRAS_NGINX_MQTT_PORT}
    ${MITRAS_NGINX_MQTTS_PORT}
    ${MITRAS_WS_ADAPTER_HTTP_PORT}' < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf

exec nginx -g "daemon off;"