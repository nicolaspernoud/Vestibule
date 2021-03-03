#!/bin/bash

###################################################
### Example production deployment with net host ###
###################################################

WD="$(
    cd "$(dirname "$0")"
    pwd -P
)"

# Tear down
docker stop vestibule
docker rm vestibule

# Variables
ROOT_DOMAIN=example.com

# Paths
mkdir -p ${WD}/letsencrypt_cache
sudo chown -Rf 1000:1000 ${WD}/letsencrypt_cache
sudo chown -Rf 1000:1000 ${WD}/configs

# Create reverse proxy
docker run -d --name vestibule \
    --net host \
    -v /etc/localtime:/etc/localtime:ro \
    -v /etc/timezone:/etc/timezone:ro \
    -v ${WD}/configs:/app/configs \
    -v ${WD}/letsencrypt_cache:/app/letsencrypt_cache \
    -e REDIRECT_URL=https://${ROOT_DOMAIN}/OAuth2Callback \
    -e CLIENT_ID=foo \
    -e CLIENT_SECRET=bar \
    -e AUTH_URL=http://localhost:8090/auth \
    -e TOKEN_URL=http://localhost:8090/token \
    -e USERINFO_URL=http://localhost:8090/admininfo \
    -e LOGOUT_URL=/ \
    -e ADMIN_ROLE=ADMINS \
    -e HOSTNAME=${ROOT_DOMAIN} \
    -e ONLYOFFICE_TITLE=VestibuleOffice \
    -e ONLYOFFICE_SERVER=https://localhost:2443 \
    -e INMEMORY_TOKEN_LIFE_DAYS=1 \
    nicolaspernoud/vestibule
