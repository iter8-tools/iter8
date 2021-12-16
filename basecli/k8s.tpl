{{/* -- id ------------------------------------------------- */}}
{{- $id := randAlphaNum 5 | lower -}}
{{- if hasKey .Values "id" -}}
  {{- $id = .Values.id -}}
{{- end -}}
{{/* -- app ------------------------------------------------ */}}
{{- $app := printf "default" -}}
{{- if hasKey .Values "app" -}}
  {{- $app = .Values.app -}}
{{- end -}}
{{/* -- loglevel ------------------------------------------- */}}
{{- $loglevel := printf "info" -}}
{{- if hasKey .Values "loglevel" -}}
  {{- $loglevel = .Values.loglevel -}}
{{- end -}}
{{/* -- name ----------------------------------------------- */}}
{{- $name := printf "experiment-%s" $id -}}
{{/* -- version -------------------------------------------- */}}
{{- $version := iter8MajorMinorVersion -}}
{{/* -- image ---------------------------------------------- */}}
{{- $image := defaultImage -}}
{{- if hasKey .Values "image" -}}
  {{- $image = .Values.image -}}
{{- end -}}
{{/* ------------------------------------------------------- */}}
{{/* -- manifest ------------------------------------------- */}}
{{/* ------------------------------------------------------- */}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ $name }}
  labels:
    app.kubernetes.io/name: iter8
    app.kubernetes.io/instance: {{ $id }}
    app.kubernetes.io/version: "{{ $version }}"
    app.kubernetes.io/component: spec
    app.kubernetes.io/created-by: iter8
    iter8.tools/app: {{ $app }}
stringData:
  experiment: |
{{ .Tasks | toYAML | indent 4 }}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{ $name }}-result
  labels:
    app.kubernetes.io/name: iter8
    app.kubernetes.io/instance: {{ $id }}
    app.kubernetes.io/version: "{{ $version }}"
    app.kubernetes.io/component: result
    app.kubernetes.io/created-by: iter8
    iter8.tools/app: {{ $app }}
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
    app.kubernetes.io/name: iter8
    app.kubernetes.io/instance: {{ $id }}
    app.kubernetes.io/version: "{{ $version }}"
    app.kubernetes.io/component: job
    app.kubernetes.io/created-by: iter8
    iter8.tools/app: {{ $app }}
spec:
  template:
    metadata:
      labels:
        app.kubernetes.io/name: iter8
        app.kubernetes.io/instance: {{ $id }}
        app.kubernetes.io/version: "{{ $version }}"
        app.kubernetes.io/component: pod
        app.kubernetes.io/created-by: iter8
        iter8.tools/app: {{ $app }}
    spec:
      containers:
      - name: iter8
        image: {{ $image }}
        imagePullPolicy: Always
        env:
        - name: LOG_LEVEL
          value: {{ $loglevel }}
        command:
        - "/bin/sh"
        - "-c"
        - iter8 k run --id {{ $id }}  # run experiment using remote secret
      restartPolicy: Never
  backoffLimit: 0
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ $name }}
  labels:
    app.kubernetes.io/name: iter8
    app.kubernetes.io/instance: {{ $id }}
    app.kubernetes.io/version: "{{ $version }}"
    app.kubernetes.io/component: rbac
    app.kubernetes.io/created-by: iter8
    iter8.tools/app: {{ $app }}
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
    app.kubernetes.io/name: iter8
    app.kubernetes.io/instance: {{ $id }}
    app.kubernetes.io/version: "{{ $version }}"
    app.kubernetes.io/component: rbac
    app.kubernetes.io/created-by: iter8
    iter8.tools/app: {{ $app }}
subjects:
- kind: ServiceAccount
  name: default
roleRef:
  kind: Role
  name: {{ $name }}
  apiGroup: rbac.authorization.k8s.io
