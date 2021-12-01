{{- $suffix := randAlphaNum 5 | lower -}}
{{- $name := printf "experiment-%s" $suffix -}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ $name }}
  labels:
    iter8/type: experiment
    iter8/experiment: {{ $name }}
stringData:
  experiment: |
{{ . | toYAML | indent 4 }}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{ $name }}-result
  labels:
    iter8/type: experiment
    iter8/experiment: {{ $name }}
stringData:
  result: |
    numCompletedTasks: 0
    failure: false
    insights: {
      metricsInfo: {}
    }
---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ $name }}
  labels:
    iter8/experiment: {{ $name }}
spec:
  template:
    spec:
      containers:
      - name: iter8
        image: kalantar/kubectl-iter8:latest
        imagePullPolicy: Always
        env:
        - name: LOG_LEVEL
          value: info
        command:
        - "/bin/sh"
        - "-c"
        - |
          set -e

          # ensure secret is created
          sleep 2
          
          # run experiment using remote secret
          kubectl iter8 run --remote -e {{ $name }}

      restartPolicy: Never
  backoffLimit: 0
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ $name }}
  labels:
    iter8/experiment: {{ $name }}
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["{{ $name }}","{{ $name }}-result"]
  verbs: ["get", "list", "patch", "create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ $name }}
  labels:
    iter8/experiment: {{ $name }}
subjects:
- kind: ServiceAccount
  name: default
roleRef:
  kind: Role
  name: {{ $name }}
  apiGroup: rbac.authorization.k8s.io
