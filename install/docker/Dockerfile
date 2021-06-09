FROM docker:dind
WORKDIR /iter8
COPY iter8.sh .
RUN chmod +x ./iter8.sh

RUN echo http://mirror.yandex.ru/mirrors/alpine/v3.5/main > /etc/apk/repositories; \
    echo http://mirror.yandex.ru/mirrors/alpine/v3.5/community >> /etc/apk/repositories

# get curl
RUN apk update && apk upgrade && apk add --no-cache curl

# get kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
RUN chmod +x ./kubectl
RUN mv ./kubectl /bin/kubectl

# get kind
RUN curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.0/kind-linux-amd64
RUN chmod +x ./kind
RUN mv ./kind /bin/kind

# get bash, git, less, openssh
RUN apk update && apk upgrade && apk add --no-cache bash git less openssh

# get kustomize
RUN curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
RUN chmod +x ./kustomize
RUN mv ./kustomize /bin/kustomize
