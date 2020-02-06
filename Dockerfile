FROM golang:stretch as basego
RUN apt-get update
RUN apt-get install -y unixodbc unixodbc-dev odbc-postgresql libsqliteodbc

FROM basego as dbconnector
WORKDIR /go/src/github.com/padwalab/dbconnector
ADD ./api /go/src/github.com/alexbrainman/odbc/api
ADD ./gosrc ./gosrc
ADD ./main.go ./main.go
RUN go build main.go 

FROM dbconnector
WORKDIR /dbconnector
COPY --from=dbconnector /go/src/github.com/padwalab/dbconnector/main .
ADD ./drivers ./drivers
ADD ./lib ./lib
CMD ["./main"]
EXPOSE 8000