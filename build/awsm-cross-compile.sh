#!/bin/bash
#
# usage: ./awsm-cross-compile.sh

env GOOS=linux GOARCH=amd64 go build -o ./linux-amd64/awsm -v github.com/murdinc/awsm
env GOOS=darwin GOARCH=amd64 go build -o ./darwin-amd64/awsm -v github.com/murdinc/awsm
env GOOS=windows GOARCH=amd64 go build -o ./windows-amd64/awsm -v github.com/murdinc/awsm