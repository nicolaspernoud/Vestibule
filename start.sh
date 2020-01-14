#!/bin/bash
. ./.env
go run main.go -apps=./configs/apps.json -davs=./configs/davs.json -https_port=1443