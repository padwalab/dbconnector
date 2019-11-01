#!/bin/sh
cd /go
go test
./app
tail -f /dev/null
