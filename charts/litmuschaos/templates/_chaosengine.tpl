{{- define "chaosengine" -}}
apiVersion: litmuschaos.io/v1alpha1
kind: ChaosEngine
metadata:
  name: {{ .Chart.Name }}-{{ .Release.Name }}
spec:
  appinfo:
    appns: "{{ .Release.Namespace }}"
    applabel: "{{ required ".Values.applabel is required!" .Values.applabel }}"
    appkind: "{{ .Values.appkind }}"
  # It can be active/stop
  engineState: 'active'
  chaosServiceAccount: {{ .Chart.Name }}-{{ .Release.Name }}
  experiments:
    - name: pod-delete
      spec:
        components:
          env:
          - name: TOTAL_CHAOS_DURATION
            value: {{ required ".Values.totalChaosDuration is required!" .Values.totalChaosDuration | quote }}
          - name: CHAOS_INTERVAL
            value: {{ required ".Values.chaosInterval is required!" .Values.chaosInterval | quote }}
{{- end }}
