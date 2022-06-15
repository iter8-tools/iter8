
Experiment summary:
*******************

  Experiment completed: {{ .Completed }}
  No task failures: {{ .NoFailure }}
  Total number of tasks: {{ len .Spec }}
  Number of completed tasks: {{ .Result.NumCompletedTasks }}

{{- if .Result.Insights }}
{{- if not (empty .Result.Insights.SLOs) }}

Whether or not service level objectives (SLOs) are satisfied:
*************************************************************

{{ .PrintSLOsText | indent 2 }}
{{- end }}

Latest observed values for metrics:
***********************************

{{ .PrintMetricsText | indent 2 }}
{{- else }}

Metrics-based Insights:
***********************

  Insights not found in experiment results. You may need to retry this report at a later time.
{{- end }}
