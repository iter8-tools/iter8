{{- define "k.job" -}}
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Release.Name }}-{{ .Release.Revision }}-job
  annotations:
    iter8.tools/revision: {{ .Release.Revision | quote }}
spec:
  template:
    metadata:
      labels:
        iter8.tools/group: {{ .Release.Name }}
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      containers:
      - name: iter8
        image: {{ .Values.iter8lib.iter8Image }}
        imagePullPolicy: Always
        command:
        - "/bin/sh"
        - "-c"
        - |
          iter8 k run --namespace {{ .Release.Namespace }} --group {{ .Release.Name }} -l {{ .Values.iter8lib.logLevel }}
      restartPolicy: Never
  backoffLimit: 0
{{- end }}
