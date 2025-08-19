# pfms
An app to Ping from multiple sources

Have you every wanted to check connectivity from different source IPs ( different routers or switches ?)

This app help you to do just that, it login's via SSH to the mentioned sources and performs ping and give you the result

All you need to do it enter the data in data.yaml and run the app

## Quick Start Guide

1. Download this repo `git clone git@github.com:JayantSeth/pfms.git`
2. Enter the repo directory `cd pfms`
3. Download the dependencies `go mod download`
4. Enter the required data in `data.yaml`
5. Run the app with `go run .`


## Supported Devices
1. Linux
2. Arista EOS
3. Cisco IOS

## Prerequisites
1. Go lang version 1.20+
