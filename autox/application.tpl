apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {{ .Name }}
  namespace: argocd
  ownerReferences:
  - apiVersion: v1
    kind: Secret
    name: {{ .Owner.Name }}
    uid: {{ .Owner.UID }}
  finalizers:
    - resources-finalizer.argocd.argoproj.io
  labels:
    app.kubernetes.io/managed-by: iter8
spec:
  destination:
    namespace: {{ .Namespace }}
    server: https://kubernetes.default.svc
  project: default
  source:
    chart: {{ .Chart.Name }}
    helm:
      values: |
        {{ .Chart.Values | toYaml | indent 8 | trim }}
    repoURL: https://iter8-tools.github.io/iter8
    targetRevision: {{ .Chart.Version }}
  ignoreDifferences:
  - kind: Secret
    name: {{ .Name }}
    namespace: {{ .Namespace }}
    jsonPointers:
    - /data
    - /metadata
  syncPolicy:
    automated:
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
    - RespectIgnoreDifferences=true