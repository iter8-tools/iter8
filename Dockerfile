# Small Linux image with Iter8 binary
FROM debian:stable-slim

# Install curl
RUN apt-get update && apt-get install -y curl

# Set Iter8 version from build args
ARG TAG
ENV TAG=${TAG:-v0.13.0}

# Download Iter8 compressed binary
RUN curl -LO https://github.com/iter8-tools/iter8/releases/download/${TAG}/iter8-linux-amd64.tar.gz

# Extract Iter8
RUN tar -xvf iter8-linux-amd64.tar.gz && rm iter8-linux-amd64.tar.gz

# Move Iter8
RUN mv linux-amd64/iter8 /bin/iter8

WORKDIR /
