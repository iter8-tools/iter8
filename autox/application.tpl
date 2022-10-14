kind: Application
metadata:
  name: {{ .name }}
  namespace: {{ .namespace }}
  ownerReferences:
    - apiVersion: {{ .ownerApiVersion }}
    kind: {{ .ownerKind}}
    name: {{ .ownerName}}
    uid: {{ .ownerUID}}
spec:
  destination:
    namespace: kubeseal
    server: https://kubernetes.default.svc
  project: default
  source:
    chart: {{ .chartName }}
    helm:
      values: 
        {{ .chartValues | toYAML | indent 4 }}
    repoURL: {{ .chartURL }}
    targetRevision: {{ .chartVersion }}
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true