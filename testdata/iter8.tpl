{{- $suffix := randAlphaNum 5 | lower -}}
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Name }}-{{ $suffix }}
spec:
  template:
    spec:
      containers:
      - name: iter8
        image: sriumcp/iter8
        command:
        - "/bin/sh"
        - "-c"
        - |
          set -e
          # trap 'kill $(jobs -p)' EXIT

          # get experiment from secret
          kubectl get secret {{ .Name }}-{{ $suffix }} -o go-template='{{"{{"}} .data.experiment {{"}}"}}' | base64 -d > experiment.yaml

          # local run
          export LOG_LEVEL=info
          iter8 run experiment.yaml

          # update the secret
          kubectl create secret generic {{ .Name }}-{{ $suffix }} --from-file=experiment=experiment.yaml --dry-run=client -o yaml | kubectl apply -f -
      restartPolicy: Never
  backoffLimit: 4
---
apiVersion: v1
kind: Secret
metadata:
  name: {{ .Name }}-{{ $suffix }}
stringData:
  experiment: |
{{ . | toYAML | indent 4 }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ .Name }}-{{ $suffix }}
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["{{ .Name }}-{{ $suffix }}"]
  verbs: ["get", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ .Name }}-{{ $suffix }}
subjects:
- kind: ServiceAccount
  name: default
roleRef:
  kind: Role
  name: {{ .Name }}-{{ $suffix }}
  apiGroup: rbac.authorization.k8s.io
