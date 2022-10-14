kind: Application
metadata:
  name: {{ .name }}
  namespace: {{ .namespace }}
  ownerReferences:
    - apiVersion: {{ .owner.apiVersion }}
    kind: Secret
    name: {{ .owner.name}}
    uid: {{ .owner.uid}}
spec:
  destination:
    namespace: kubeseal
    server: https://kubernetes.default.svc
  project: default
  source:
    chart: {{ .chart.name }}
    helm:
      values: 
        {{ .chart.values | toYAML | indent 4 }}
    repoURL: {{ .chart.url }}
    targetRevision: {{ .chart.version }}
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true