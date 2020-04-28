#!/bin/bash

# This script does the following :
# 1. Packages the ../lambda/main.go file in a structure that is consumable by AWS Lambda.
# 2. Saves to temporary local location

env GOOS=linux GOARCH=amd64 go build -o /tmp/main $1
zip -j /tmp/main.zip /tmp/main
