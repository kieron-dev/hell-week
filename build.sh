#!/bin/bash

export GOOS=linux
export GOARCH=amd64

go build -v -o hell github.com/kieron-pivotal/container-run
go build -v -o pin-cpu github.com/kieron-pivotal/container-run/pincpu 
