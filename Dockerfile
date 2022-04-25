# Get Iter8
FROM golang:1.17-buster as builder

# Install wget
RUN apt-get update && apt-get install -y wget

# Set Iter8 version from build args
ARG TAG
ENV TAG=${TAG:-v0.10.6}

# Download iter8 compressed binary
RUN wget https://github.com/iter8-tools/iter8/releases/download/${TAG}/iter8-linux-amd64.tar.gz

# Extract iter8
RUN tar -xvf iter8-linux-amd64.tar.gz

# Move iter8
RUN mv linux-amd64/iter8 /bin/iter8

# Small linux image with iter8 binary
FROM debian:buster-slim
WORKDIR /
COPY --from=builder /bin/iter8 /bin/iter8
