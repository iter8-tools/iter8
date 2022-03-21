{{- define "k8sjob" -}}
{{- $name := printf "%v-%v" .Release.Name .Release.Revision -}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ $name }}-spec
stringData:
  experiment.yaml: |
{{ include "experiment" . | indent 4 }}
---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ $name }}-job
spec:
  template:
    spec:
      containers:
      - name: iter8
        image: iter8-tools/iter8:{{  trimPrefix "v" .Chart.AppVersion }}
        imagePullPolicy: Always
        volumeMounts:
        - name: experiment-spec
          mountPath: "/iter8exp"
          readOnly: true        
        command:
        - "/bin/sh"
        - "-c"
        - |
          cd /iter8exp \
          iter8 k run --namespace {{ .Release.Namespace }} --group {{ .Release.Name }} --revision {{ .Release.Revision }}
      volumes:
      - name: experiment-spec
        secret:
          secretName: {{ $name }}-spec
      restartPolicy: Never
  backoffLimit: 0
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ $name }}
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["{{ $name }}-result"]
  verbs: ["get", "list", "patch", "update", "create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ $name }}
subjects:
- kind: ServiceAccount
  name: default
roleRef:
  kind: Role
  name: {{ $name }}
  apiGroup: rbac.authorization.k8s.io
{{- end -}}