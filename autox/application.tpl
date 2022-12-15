apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {{ .Name }}
  namespace: argocd
  ownerReferences:
  - apiVersion: v1
    kind: Secret
    name: {{ .Owner.Name }}
    uid: {{ .Owner.Uid }}
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
    repoURL: https://iter8-tools.github.io/hub
    targetRevision: {{ .Chart.Version }}
  syncPolicy:
    automated:
      selfHeal: false
    syncOptions:
    - CreateNamespace=true