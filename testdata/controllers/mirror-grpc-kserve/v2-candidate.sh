cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: candidate
---
apiVersion: v1
kind: Service
metadata:
  name: wisdom-mirror
  namespace: candidate
spec:
  clusterIP: None
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
---
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: wisdom
  namespace: candidate
  annotations:
    proxy.istio.io/config: '{"holdApplicationUntilProxyStarts": true}'
  labels:
    app.kubernetes.io/name: wisdom
    app.kubernetes.io/version: v2
    iter8.tools/watch: "true"
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
---
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: wisdom-mirror
  namespace: candidate
spec:
  gateways:
  - knative-serving/knative-local-gateway
  hosts:
  - wisdom-mirror.candidate
  - wisdom-mirror.candidate.svc.cluster.local
  http:
  - route:
    - destination:
        host: knative-local-gateway.istio-system.svc.cluster.local
      headers:
        request:
          set:
            Host: wisdom-predictor-default.candidate.svc.cluster.local
EOF