
Experiment summary:
*******************

  Experiment completed: {{ .Completed }}
  No task failures: {{ .NoFailure }}
  Total number of tasks: {{ len .Tasks }}
  Number of completed tasks: {{ .Result.NumCompletedTasks }}

{{- if not (empty .Result.Insights.SLOs) }}

Whether or not service level objectives (SLOs) are satisfied:
*************************************************************

{{ .PrintSLOsText | indent 2 }}
{{- end }}

Latest observed values for metrics:
***********************************

{{ .PrintMetricsText | indent 2 }}