#!/bin/sh
go build -o main main.go
docker build -t gojs .
