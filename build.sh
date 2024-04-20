#!/bin/sh
BUILD_TIME=$(date)
FLAG="-X main.buildTime=$BUILD_TIME"

go build -v -ldflags "$FLAG"