##
## Build
##
# NOTE: make sure to specify the platform in case we build on M1 macOS
# NOTE: make sure the base image lists amd64 on Docker Hub page! https://hub.docker.com/_/golang/tags
FROM --platform=linux/amd64 golang:1.21-alpine AS build

RUN apk update && apk add gcc musl-dev

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# overwrite this at build time
# put this here so previous layers do not get invalidated
# https://stackoverflow.com/questions/60450479/using-arg-and-env-in-dockerfile
ARG Version=foo-docker-version

COPY main.go ./
# RUN go test -v ./...
RUN go build -ldflags="-X 'main.Version=$Version'" -o /fastqSplit main.go

##
## Deploy
##

# need alpine for using bash, otherwise use scratch
# https://hub.docker.com/_/alpine/tags
# FROM --platform=linux/amd64 alpine:3.18.4
# RUN apk add bash

# NOTE: had issues with alpine on AWS Batch so use Debian or Ubuntu instead
FROM --platform=linux/amd64 ubuntu:22.04
RUN apt update -y && apt upgrade -y && apt install -y pigz

# Also consider Debian Slim
# https://www.nextflow.io/docs/latest/tracing.html#trace-required-packages
# FROM --platform=linux/amd64 debian:bookworm-slim
# RUN apt update -y && apt upgrade -y && apt install -y procps

COPY --from=build /fastqSplit /usr/local/bin/fastqSplit
RUN ln -s /usr/local/bin/fastqSplit
RUN which fastqSplit
RUN fastqSplit -h