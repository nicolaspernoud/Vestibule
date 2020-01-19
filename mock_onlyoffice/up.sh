#!/bin/bash
WD="$(
    cd "$(dirname "$0")"
    pwd -P
)"
$WD/down.sh
docker run -d --name onlyoffice \
    --restart unless-stopped \
    -v /etc/localtime:/etc/localtime:ro \
    -v /etc/timezone:/etc/timezone:ro \
    -v $WD/data:/var/www/onlyoffice/Data \
    -p 2443:443 \
    -e "DOCKER_HOST=$(ip -4 addr show docker0 | grep -Po 'inet \K[\d.]+')" \
    onlyoffice/documentserver

docker exec -it onlyoffice /var/www/onlyoffice/documentserver/npm/node_modules/.bin/json -f /etc/onlyoffice/documentserver/default.json -I -e 'this.services.CoAuthoring.requestDefaults.rejectUnauthorized=false'
docker exec -it onlyoffice /bin/bash -c 'echo "$DOCKER_HOST vestibule.127.0.0.1.nip.io" >>/etc/hosts'
docker exec -it onlyoffice /bin/bash -c 'echo "$DOCKER_HOST unsecureddav.vestibule.127.0.0.1.nip.io" >>/etc/hosts'
docker exec -it onlyoffice /bin/bash -c 'echo "$DOCKER_HOST userdav.vestibule.127.0.0.1.nip.io" >>/etc/hosts'
docker exec -it onlyoffice /bin/bash -c 'echo "$DOCKER_HOST admindav.vestibule.127.0.0.1.nip.io" >>/etc/hosts'
