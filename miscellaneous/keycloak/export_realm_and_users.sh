#!/bin/bash
docker exec -it keycloak keycloak/bin/standalone.sh \
    -Djboss.socket.binding.port-offset=100 \
    -Dkeycloak.migration.action=export \
    -Dkeycloak.migration.provider=singleFile \
    -Dkeycloak.migration.realmName=mdl \
    -Dkeycloak.migration.usersExportStrategy=REALM_FILE \
    -Dkeycloak.migration.file=/tmp/kc/mdl-realm.json
