FROM golang:latest as builder
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN go build -tags go_tarantool_ssl_disable ./cmd/main.go
EXPOSE 1234

RUN mkdir -p /var/app/configs
VOLUME /var/app/configs
CMD ["./main", "-config", "/var/app/configs/config.yaml"]
