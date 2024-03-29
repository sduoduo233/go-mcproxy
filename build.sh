#!/bin/bash
export GOARCH=amd64
export CGO_ENABLED=0

export GOOS=windows
go build -o ./build/mcproxy.exe -trimpath -ldflags "-s -w"

export GOOS=linux
go build -o ./build/mcproxy -trimpath -ldflags "-s -w"