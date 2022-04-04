{{- define "k.job" -}}
{{- $name := printf "%v-%v" .Release.Name .Release.Revision -}}
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
        command:
        - "/bin/sh"
        - "-c"
        - |
          iter8 k run --namespace {{ .Release.Namespace }} --group {{ .Release.Name }} --revision {{ .Release.Revision }}
      restartPolicy: Never
  backoffLimit: 0
{{- end }}
