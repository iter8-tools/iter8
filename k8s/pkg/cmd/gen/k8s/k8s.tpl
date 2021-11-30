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
        image: kalantar/kubectl-iter8:20211124-1640
        # imagePullPolicy: Always
        command:
        - "/bin/sh"
        - "-c"
        - |
          set -e
          # trap 'kill $(jobs -p)' EXIT

          # get experiment from secret
          sleep 2 # let secret be created
          echo getting secret {{ $name }}
          kubectl get secret {{ $name }} -o jsonpath='{.data.experiment}' | base64 -d > experiment.yaml
          
          # local run
          export LOG_LEVEL=info
          kubectl-iter8 run

          # update the secret
          kubectl create secret generic {{ $name }}-result --from-file=result=result.yaml --dry-run=client -o yaml | kubectl apply -f -
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
# ---
# apiVersion: rbac.authorization.k8s.io/v1
# kind: Role
# metadata:
#   name: {{ $name }}-pods
#   labels:
#     iter8/experiment: {{ $name }}
# rules:
# - apiGroups: [""]
#   resources: ["pods","pods/log","secrets"]
#   verbs: ["get", "list", "patch", "create"]
# - apiGroups: ["batch"]
#   resources: ["jobs"]
#   verbs: ["get", "list", "patch", "create"]
# ---
# apiVersion: rbac.authorization.k8s.io/v1
# kind: RoleBinding
# metadata:
#   name: {{ $name }}
#   labels:
#     iter8/experiment: {{ $name }}
# subjects:
# - kind: ServiceAccount
#   name: default
# roleRef:
#   kind: Role
#   name: {{ $name }}-pods
#   apiGroup: rbac.authorization.k8s.io
