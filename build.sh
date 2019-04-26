#!/bin/bash

export GOOS=linux
export GOARCH=amd64

go build -v -o hell github.com/kieron-pivotal/hell-week/container-run
go build -v -o container-runc-run github.com/kieron-pivotal/hell-week/container-runc
go build -v -o pin-cpu github.com/kieron-pivotal/hell-week/pincpu
go build -v -o limit-memory github.com/kieron-pivotal/hell-week/limitmem
go build -v -o rootfs-make github.com/kieron-pivotal/hell-week/rootfsmake
