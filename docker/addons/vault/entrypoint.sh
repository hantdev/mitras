#!/usr/bin/dumb-init /bin/sh

VAULT_CONFIG_DIR=/vault/config

docker-entrypoint.sh server &
VAULT_PID=$!

sleep 2

echo $MITRAS_VAULT_UNSEAL_KEY_1
echo $MITRAS_VAULT_UNSEAL_KEY_2
echo $MITRAS_VAULT_UNSEAL_KEY_3

if [[ ! -z "${MITRAS_VAULT_UNSEAL_KEY_1}" ]] &&
   [[ ! -z "${MITRAS_VAULT_UNSEAL_KEY_2}" ]] &&
   [[ ! -z "${MITRAS_VAULT_UNSEAL_KEY_3}" ]]; then
	echo "Unsealing Vault"
	vault operator unseal ${MITRAS_VAULT_UNSEAL_KEY_1}
	vault operator unseal ${MITRAS_VAULT_UNSEAL_KEY_2}
	vault operator unseal ${MITRAS_VAULT_UNSEAL_KEY_3}
fi

wait $VAULT_PID