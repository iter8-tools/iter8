cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: stable
---
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "wisdom"
  namespace: stable
  annotations:
    proxy.istio.io/config: '{"holdApplicationUntilProxyStarts": true}'
spec:
  predictor:
    model:
      modelFormat:
        name: sklearn
      runtime: kserve-mlserver
      storageUri: "gs://seldon-models/sklearn/mms/lr_model"
      ports:
      - containerPort: 9000
        name: h2c
        protocol: TCP      
EOF