#!/bin/sh
echo $PWD
cd /go/src/github.com/alexbrainman/odbc/
go test -mssrv=52.53.245.117 -msdb=northwind -msuser=admin -mspass=Tibco123 -v -run=MS -msdriver="ODBC Driver 17 for SQL Server"
