FROM debian:stretch-slim
WORKDIR /dbconnector
RUN apt-get update
RUN apt-get install -y unixodbc unixodbc-dev odbc-postgresql libsqliteodbc

ADD ./main .
ADD ./drivers ./drivers
ADD ./lib ./lib
CMD ["./main"]
EXPOSE 8000