#!/usr/bin/bash

vault() {
    if is_container_running "mitras-vault"; then
        docker exec -it mitras-vault vault "$@"
    else
        if which vault &> /dev/null; then
            $(which vault) "$@"
        else
            echo "mitras-vault container or vault command not found."
        fi
    fi
}

is_container_running() {
    local container_name="$1"
    if [ "$(docker inspect --format '{{.State.Running}}' "$container_name" 2>/dev/null)" = "true" ]; then
        return 0
    else
        return 1
    fi
}