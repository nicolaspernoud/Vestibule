#!/bin/bash
export $(cat .env | xargs)
go run main.go
