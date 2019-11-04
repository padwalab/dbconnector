FROM ubuntu:latest

RUN apt-get update && apt-get install -y \
    curl \
    unixodbc \
    unixodbc-dev \
    odbc-postgresql

RUN rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.4.2

RUN curl -sSL https://storage.googleapis.com/golang/go$GOLANG_VERSION.linux-amd64.tar.gz \
    | tar -v -C /usr/local -xz

ENV PATH /usr/local/go/bin:$PATH

RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
WORKDIR /go

RUN mkdir app
COPY ./app ./app
# COPY ./script.sh .

# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .
# RUN go build -a -o app .
# RUN go test
# RUN odbcinst -q -d

RUN chmod 777 ./app/cTest.sh
CMD "./app/cTest.sh"