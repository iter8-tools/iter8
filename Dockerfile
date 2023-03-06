# Golang linux image with iter8 binary
FROM golang:1.20.1-alpine

# Set Iter8 version from build args
ARG TAG
ENV TAG=${TAG:-v0.13.8}

# Download iter8 compressed binary
RUN go install github.com/iter8-tools/iter8@${TAG}

# Download curl
RUN apk --no-cache add curl

WORKDIR /
