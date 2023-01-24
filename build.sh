#!/bin/bash
export GOARCH=amd64

export GOOS=windows
go build -o ./build/mcproxy.exe 

export GOOS=linux
go build -o ./build/mcproxy