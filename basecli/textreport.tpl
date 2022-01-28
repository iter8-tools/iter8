
Experiment summary:
*******************

  Experiment completed: {{ .Completed }}
  No failed tasks: {{ .NoFailure }}
  Total number of tasks: {{ len .Tasks }}
  Number of completed tasks: {{ .Result.NumCompletedTasks }}

Whether or not service level objectives (SLOs) are satisfied:
*************************************************************

{{ .PrintSLOsText | indent 2 }}

Latest observed values for metrics:
***********************************

{{ .PrintMetricsText | indent 2 }}