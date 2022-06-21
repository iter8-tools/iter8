{{- define "k.job" -}}
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Release.Name }}-{{ .Release.Revision }}-job
  annotations:
    iter8.tools/group: {{ .Release.Name }}
    iter8.tools/revision: {{ .Release.Revision | quote }}
spec:
  template:
    metadata:
      labels:
        iter8.tools/group: {{ .Release.Name }}
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccountName: {{ .Release.Name }}-iter8-sa
      containers:
      - name: iter8
        image: {{ .Values.iter8Image }}
        imagePullPolicy: Always
        command:
        - "/bin/sh"
        - "-c"
        - |
          iter8 k run --namespace {{ .Release.Namespace }} --group {{ .Release.Name }} -l {{ .Values.logLevel }}
      restartPolicy: Never
  backoffLimit: 0
{{- end }}
