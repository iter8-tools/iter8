# Get Iter8
FROM golang:1.18-buster as builder

WORKDIR /

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY abn/ abn/
COPY base/ base/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o iter8-abn abn/main.go

### Multi-stage Docker build
### New image below

# Small linux image with iter8 binary
FROM debian:buster-slim
WORKDIR /
COPY --from=builder /iter8-abn /bin/iter8-abn

ENTRYPOINT ["/bin/iter8-abn"]

# export ABN_SERVICE_TAG=kalantar/abn-service:20220616-1415
# docker build -t $ABN_SERVICE_TAG -f abn/Dockerfile .
# docker push $ABN_SERVICE_TAG

# helm install abn ../hub/charts/abn-service/ --set image=$ABN_SERVICE_TAG --set resources='{deployments,services}' --set namespaces='{default}'
# helm upgrade abn ../hub/charts/abn-service/ --set image=$ABN_SERVICE_TAG --set resources='{deployments,services}' --set namespaces='{default}'
# k port-forward svc/abn 50051:50051