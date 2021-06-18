FROM docker:dind
 
#install go
COPY --from=golang:1.15-alpine /usr/local/go/ /usr/local/go/
 
ENV PATH="/usr/local/go/bin:${PATH}"

RUN echo http://mirror.yandex.ru/mirrors/alpine/v3.5/main > /etc/apk/repositories; \
    echo http://mirror.yandex.ru/mirrors/alpine/v3.5/community >> /etc/apk/repositories

# get curl
RUN apk update && apk upgrade && apk add --no-cache curl

# get bash, git, less, openssh
RUN apk update && apk upgrade && apk add --no-cache bash git less openssh

# get helm
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
RUN chmod 700 get_helm.sh
RUN ./get_helm.sh

# get kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
RUN chmod +x ./kubectl
RUN mv ./kubectl /bin/kubectl

# get kind
RUN curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.0/kind-linux-amd64
RUN chmod +x ./kind
RUN mv ./kind /bin/kind

# get kustomize
RUN curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
RUN chmod +x ./kustomize
RUN mv ./kustomize /bin/kustomize

# get iter8ctl
RUN GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl@v0.1.4

# clone Iter8 repo
RUN git clone https://github.com/iter8-tools/iter8
WORKDIR /iter8/install/docker
