#!/bin/sh
cd /go/app
gcc odbcFunc.c -o main.out -lodbc
./main.out
