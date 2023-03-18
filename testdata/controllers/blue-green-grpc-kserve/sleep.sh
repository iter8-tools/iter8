#!/bin/sh
cat <<EOF | istioctl kube-inject -f - | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sleep
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sleep
  template:
    metadata:
      labels:
        app: sleep
      annotations:
        inject.istio.io/templates: grpc-agent
    spec:
      containers:
      - name: sleep
        image: fullstorydev/grpcurl:latest-alpine
        command: ["/bin/sh", "-c", "sleep infinity"]
        imagePullPolicy: IfNotPresent
EOF

# Use this pod as follows
# First get $SLEEP_POD
### SLEEP_POD=$(kubectl get pod -l app=sleep -o jsonpath={.items..metadata.name})
# Second, exec into it
### kubectl exec --stdin --tty "${SLEEP_POD}" -c sleep -- /bin/sh
# Third, download protofile
### cd tmp
### wget https://raw.githubusercontent.com/kserve/kserve/master/docs/predict-api/v2/grpc_predict_v2.proto
# Fourth, download json input
### wget https://gist.githubusercontent.com/kalantar/6e9eaa03cad8f4e86b20eeb712efef45/raw/56496ed5fa9078b8c9cdad590d275ab93beaaee4/sklearn-irisv2-input-grpc.json
# Fifth, edit json file so that it has the correct model name, wisdom
### mv sklearn-irisv2-input-grpc.json wisdom.json # and vi and edit and save
# Sixth, call wisdom-greatest
### cat wisdom.json | grpcurl -plaintext -proto grpc_predict_v2.proto -d @ greatest-wisdom.default:80 inference.GRPCInferenceService.ModelInfer

