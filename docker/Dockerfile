# FROM scratch
# ADD iter8 /bin
# ENTRYPOINT ["/bin/iter8"]
FROM debian:buster-slim

# Install curl
RUN apt-get update && apt-get install -y curl

# Set Iter8 version from build args
ARG TAG
ENV TAG=${TAG:-v0.17.0}

ADD iter8 /bin

WORKDIR /
