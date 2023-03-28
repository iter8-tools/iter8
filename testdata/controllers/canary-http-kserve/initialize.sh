cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: wisdom
spec:
  ports:
  - protocol: TCP
    port: 80
---
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: wisdom-primary
  labels:
    app.kubernetes.io/name: wisdom
    app.kubernetes.io/version: v1
    iter8.tools/watch: "true"
spec:
  predictor:
    minReplicas: 1
    model:
      modelFormat:
        name: sklearn
      runtime: kserve-mlserver
      storageUri: "gs://seldon-models/sklearn/mms/lr_model"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: wisdom
  labels:
    app.kubernetes.io/managed-by: iter8
    iter8.tools/kind: subject
    iter8.tools/version: v0.14
data:
  strSpec: |
    variants: 
    - resources:
      - gvrShort: isvc
        name: wisdom-primary
    - resources:
      - gvrShort: isvc
        name: wisdom-candidate
    # routing templates
    ssas:
      canary-wisdom:
        gvrShort: vs
        template: |
          apiVersion: networking.istio.io/v1beta1
          kind: VirtualService
          metadata:
            name: wisdom
          spec:
            gateways:
            - mesh
            hosts:
            - wisdom
            - wisdom.default
            - wisdom.default.svc.cluster.local
            http:
            {{- if gt (index .Weights 1) 0 }}
            - name: green
              match:
              - headers:
                  traffic: 
                    exact: test
                uri:
                  prefix: /enlightenme
              rewrite:
                uri: /v2/models/wisdom-candidate/infer
                authority: wisdom-candidate-predictor-default.default.svc.cluster.local
              route:
              - destination:
                  host: wisdom-candidate-predictor-default.default.svc.cluster.local
            {{- end }}
            - name: blue
              match:
              - uri:
                  prefix: /enlightenme
              rewrite:
                uri: /v2/models/wisdom-primary/infer
                authority: wisdom-primary-predictor-default.default.svc.cluster.local
              route:
              - destination:
                  host: wisdom-primary-predictor-default.default.svc.cluster.local
immutable: true            
EOF