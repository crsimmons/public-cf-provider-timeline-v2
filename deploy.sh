#!/bin/bash

set -eu

GOOS=linux GOARCH=amd64 go build -o app main.go

cf push --vars-file secrets.yml
