FROM golang:1.21.6-bookworm AS build-stage

WORKDIR /app
COPY . ./

RUN go mod download

RUN mkdir -p bin \
    && make clean \
    && make build


FROM debian:bookworm-slim 

WORKDIR /

# Install curl
RUN apt-get update && apt-get install -y curl

# Install /bin/iter8
COPY --from=build-stage /app/bin/iter8 /bin/iter8

# Set Iter8 version from build args
ARG TAG
ENV TAG=${TAG:-v1.1.0}

