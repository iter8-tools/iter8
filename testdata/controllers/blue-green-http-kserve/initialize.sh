cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: wisdom
spec:
  externalName: knative-local-gateway.istio-system.svc.cluster.local
  sessionAffinity: None
  type: ExternalName
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
    iter8.tools/kind: routemap
    iter8.tools/version: v0.14
data:
  strSpec: |
    versions: 
    - weight: 60
      resources:
      - gvrShort: isvc
        name: wisdom-primary
    - weight: 40
      resources:
      - gvrShort: isvc
        name: wisdom-candidate
    # routing templates
    routingTemplates:
      blue-green-wisdom:
        gvrShort: vs
        template: |
          apiVersion: networking.istio.io/v1beta1
          kind: VirtualService
          metadata:
            name: wisdom
          spec:
            gateways:
            - mesh
            - knative-serving/knative-ingress-gateway
            - knative-serving/knative-local-gateway
            hosts:
            - wisdom
            - wisdom.default
            - wisdom.default.svc.cluster.local
            http:
            - name: blue
              match:
              - uri:
                  prefix: /enlightenme
              {{- if gt (index .Weights 1) 0 }}
                headers:
                  branch: 
                    exact: blue
              {{- end }}
              rewrite:
                uri: /v2/models/wisdom-primary/infer
              route:
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                headers:
                  request:
                    set:
                      Host: wisdom-primary-predictor-default.default.svc.cluster.local
                    {{- if gt (index .Weights 1) 0 }}
                    remove:
                    - branch
                    {{- end }}
            {{- if gt (index .Weights 1) 0 }}
            - name: green
              match:
              - headers:
                  branch: 
                    exact: green
                uri:
                  prefix: /enlightenme
              rewrite:
                uri: /v2/models/wisdom-candidate/infer
              route:
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                headers:
                  request:
                    set:
                      Host: wisdom-candidate-predictor-default.default.svc.cluster.local
                    remove:
                    - branch
            - name: split
              match:
              - uri:
                  prefix: /enlightenme
              route:
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                weight: {{ index .Weights 0 }}
                headers:
                  request:
                    set:
                      branch: blue
                      Host: wisdom.default
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                weight: {{ index .Weights 1 }}
                headers:
                  request:
                    set:
                      branch: green
                      Host: wisdom.default
            {{- end }}
immutable: true            
EOF