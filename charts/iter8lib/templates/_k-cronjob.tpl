{{- define "k.cronjob" -}}
apiVersion: batch/v1
kind: CronJob
metadata:
  name: {{ .Release.Name }}-{{ .Release.Revision }}-cronjob
  annotations:
    iter8.tools/revision: {{ .Release.Revision | quote }}
spec:
  schedule: "{{ .Values.cronjobSchedule }}"
  concurrencyPolicy: Forbid
  jobTemplate:
    spec:
      template:
        metadata:
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
              iter8 k run --namespace {{ .Release.Namespace }} --group {{ .Release.Name }} --reuseResult -l {{ .Values.iter8lib.logLevel }}
          restartPolicy: Never
      backoffLimit: 0
{{- end }}
