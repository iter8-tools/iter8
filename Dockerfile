# Small Linux image with Iter8 binary
FROM golang:buster

# Install curl
RUN apt-get update && apt-get install -y golang-go

# Install Iter8
RUN go install github.com/iter8-tools/iter8@v0.14

WORKDIR /
