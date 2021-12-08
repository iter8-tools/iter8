# Build Iter8
FROM golang:1.16-buster as builder

WORKDIR /workspace

# Copy the go source
COPY ./ ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Build
WORKDIR /workspace
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o /bin/iter8 main.go
WORKDIR /workspace

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
RUN chmod 755 kubectl
RUN cp kubectl /bin

# Install Helm 3
RUN curl -fsSL -o helm-v3.5.0-linux-amd64.tar.gz https://get.helm.sh/helm-v3.5.0-linux-amd64.tar.gz
RUN tar -zxvf helm-v3.5.0-linux-amd64.tar.gz
RUN linux-amd64/helm version

# Install Kustomize v3
RUN curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
RUN cp kustomize /bin

# Install yq
RUN GO111MODULE=on GOBIN=/bin go get github.com/mikefarah/yq/v4

# Small linux image with useful shell commands
FROM debian:buster-slim
WORKDIR /
COPY --from=builder /bin/iter8 /bin/iter8
COPY --from=builder /bin/kubectl /bin/kubectl
COPY --from=builder /bin/kustomize /bin/kustomize
COPY --from=builder /workspace/linux-amd64/helm /bin/helm
COPY --from=builder /bin/yq /bin/yq

# Install git
RUN apt-get update && apt-get install -y git curl gpg

# Install GH CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg
RUN echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null
RUN apt update && apt install gh
