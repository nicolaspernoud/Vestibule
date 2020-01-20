#!/bin/bash
# Working directory
WD="$(
    cd "$(dirname "$0")"
    pwd -P
)"

# STOP KEYCLOAK
$WD/keycloak-down.sh

# START KEYCLOAK
docker run -d --name keycloak \
    --restart unless-stopped \
    -v /etc/localtime:/etc/localtime:ro \
    -v /etc/timezone:/etc/timezone:ro \
    -p 8080:8080 \
    -e KEYCLOAK_USER=admin \
    -e KEYCLOAK_PASSWORD=admin \
    -e KEYCLOAK_IMPORT=/tmp/kc/mdl-realm.json \
    -v $WD:/tmp/kc \
    jboss/keycloak