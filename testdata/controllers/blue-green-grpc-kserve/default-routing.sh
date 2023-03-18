#!/bin/sh
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: greatest-wisdom
spec:
  externalName: istio-ingressgateway.istio-system.svc.cluster.local
  sessionAffinity: None
  type: ExternalName
---
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: greatest-wisdom
spec:
  gateways:
  - mesh
  - knative-serving/knative-ingress-gateway
  - knative-serving/knative-local-gateway
  hosts:
  - greatest-wisdom
  - greatest-wisdom.default
  - greatest-wisdom.default.svc.cluster.local
  http:
  - name: blue
    match:
    - headers:
        branch:
          exact: blue
    rewrite:
      authority: wisdom.stable.example.com
    route:
    - destination:
        host: istio-ingressgateway.istio-system.svc.cluster.local
  - name: green
    match:
    - headers:
        branch:
          exact: green
    rewrite:
      authority: wisdom.candidate.example.com
    route:
    - destination:
        host: istio-ingressgateway.istio-system.svc.cluster.local
  - name: split
    route:
    - destination:
        host: istio-ingressgateway.istio-system.svc.cluster.local
      weight: 50
      headers:
        request:
          set:
            branch: blue
            Host: greatest-wisdom.default.svc.cluster.local
    - destination:
        host: istio-ingressgateway.istio-system.svc.cluster.local
      weight: 50
      headers:
        request:
          set:
            branch: green
            Host: greatest-wisdom.default.svc.cluster.local
EOF
