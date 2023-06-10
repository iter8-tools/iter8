# Small linux image with iter8 binary
FROM debian:buster-slim

# Install curl
RUN apt-get update && apt-get install -y curl

# Set Iter8 version from build args
ARG TAG
ENV TAG=${TAG:-v0.14.10}

# Download iter8 compressed binary
RUN curl -LO https://github.com/iter8-tools/iter8/releases/download/${TAG}/iter8-linux-amd64.tar.gz

# Extract iter8
RUN tar -xvf iter8-linux-amd64.tar.gz && rm iter8-linux-amd64.tar.gz

# Move iter8
RUN mv linux-amd64/iter8 /bin/iter8

WORKDIR /
