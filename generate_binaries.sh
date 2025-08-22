#!/bin/bash

GOOS=windows GOARCH=amd64 go build -o pfms-windows-amd64.exe .

GOOS=darwin GOARCH=amd64 go build -o pfms-mac-amd64 .

GOOS=linux GOARCH=amd64 go build -o pfms-linux-amd64 .


