{{- define "metrics.common.linkerd" }}        
        {{"{{"}}- if .Values.deployment {{"}}"}}
          deployment="{{"{{"}}.deployment{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .Values.namespace {{"}}"}}
          namespace="{{"{{"}}.namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
{{- end }}