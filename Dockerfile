# Start from the latest golang base image
#FROM golang:latest as builder
FROM golang:alpine AS build

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o ./out/xdscli .

FROM alpine
WORKDIR /app
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=build /build/out/xdscli /app/
ENTRYPOINT [ "./xdscli" ]