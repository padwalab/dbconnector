FROM debian:stretch as basegojs
RUN apt-get update
RUN apt-get install -y unixodbc unixodbc-dev odbc-postgresql libsqliteodbc

FROM basegojs
WORKDIR /gojs
ADD ./drivers ./drivers
ADD   ./main .
ADD ./lib ./lib


CMD ["/gojs/main"]
EXPOSE 8000