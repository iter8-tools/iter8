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
        image: iter8/iter8
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
    # task 1: generate HTTP requests for https://example.com
    # collect Iter8's built-in latency and error-related metrics
    - task: gen-load-and-collect-metrics
      with:
        versionInfo:
        - url: https://example.com
    # task 2: validate service level objectives for https://example.com using
    # the metrics collected in the above task
    - task: assess-app-versions
      with:
        SLOs:
          # error rate must be 0
        - metric: built-in/error-rate
          upperLimit: 0
          # 95th percentile latency must be under 1 msec
        - metric: built-in/p95.0
          upperLimit: 1
    # task 3: if SLOs are satisfied, do something
    - if: SLOs()
      run: echo "SLOs are satisfied"
    # task 4: if SLOs are not satisfied, do something else
    - if: not SLOs()
      run: echo "SLOs are not satisfied"
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
