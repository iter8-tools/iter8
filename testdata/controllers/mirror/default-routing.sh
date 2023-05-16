#!/bin/sh
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: httpbin
  labels:
    app: httpbin
spec:
  ports:
  - name: http
    port: 8000
    targetPort: 80
  selector:
    app: httpbin
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: httpbin
spec:
  hosts:
    - httpbin
  http:
  - route:
    - destination:
        host: httpbin
        subset: v1
      weight: 100
---
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: httpbin
spec:
  host: httpbin
  subsets:
  - name: v1
    labels:
      version: v1
  - name: v2
    labels:
      version: v2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: httpbin
  labels:
    app.kubernetes.io/managed-by: iter8
    iter8.tools/kind: routemap
    iter8.tools/version: v0.14        
data:
  strSpec: |
    versions: 
    # version 1 is left unspecified since it does not have a role in routing
    - 
    - resources:
      - gvrShort: deploy
        name: httpbin-v2
    routingTemplates:
      # templates for server-side apply of resources
      mirror-httpbin:
        gvrShort: vs
        template: |
          apiVersion: networking.istio.io/v1beta1
          kind: VirtualService
          metadata:
            name: httpbin
          spec:
            http:
            - route:
              - destination:
                  host: httpbin
                  subset: v1
                weight: 100
              {{- if gt (index .Weights 1) 0 }}
              mirror:
                host: httpbin
                subset: v2
              mirrorPercentage:
                value: 100.0
              {{- end }}
immutable: true    
EOF
