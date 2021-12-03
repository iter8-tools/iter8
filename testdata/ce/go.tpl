{{- $suffix := randAlphaNum 5 | lower -}}
apiVersion: codeengine.cloud.ibm.com/v1beta1
kind: JobRun
metadata:
  name: {{ .Name }}-{{ $suffix }}
spec:
  jobDefinitionSpec:
    template:
      containers:
      - name: iter8
        image: puffinmuffin/iter8
        command:
        - "/bin/sh"
        - "-c"
        - |
          set -e
          # trap 'kill $(jobs -p)' EXIT

          # get experiment from secret
          kubectl get secret {{ .Name }}-{{ $suffix }} -o go-template='{{"{{"}} .data.experiment {{"}}"}}' | base64 -d > experiment.yaml

          # local run
          export LOG_LEVEL=info
          iter8 run experiment.yaml

          # update the secret
          kubectl create secret generic {{ .Name }}-{{ $suffix }} --from-file=experiment=experiment.yaml --dry-run=client -o yaml | kubectl apply -f -
---
apiVersion: v1
kind: Secret
metadata:
  name: {{ .Name }}-{{ $suffix }}
stringData:
  experiment: |
    # task 1: generate HTTP requests for https://example.com
    # collect Iter8's built-in latency and error-related metrics
    - task: gen-load-and-collect-metrics
      with:
        versionInfo:
        - url: https://iter8-ce-demo.fgw1ut94lpp.ca-tor.codeengine.appdomain.cloud
    # task 2: validate service level objectives for https://example.com using
    # the metrics collected in the above task
    - task: assess-app-versions
      with:
        SLOs:
          # error rate must be 0
        - metric: built-in/error-rate
          upperLimit: 0
          # 95th percentile latency must be under 1 msec
        - metric: built-in/p95.0
          upperLimit: 1
    # task 3: if SLOs are satisfied, do nothing
    - if: SLOs()
      run: echo "SLOs are satisfied"
    # task 4: if SLOs are not satisfied, revert to the previous version
    - if: not SLOs()
      run: |
        # Get the name of the second to last revision
        penultimateRevisionName=$(kubectl get revisions \
        -l serving.knative.dev/configuration={{ .CEAppName }} \
        -o go-template='{{"{{"}}range .items{{"}}"}} {{"{{"}}index . "metadata" "name"{{"}}"}} {{"{{"}}end{{"}}"}}' \
        | awk '{print $(NF - 1)}'); \
        
        # Shift all the traffic weight away from the last revision to the second last
        kubectl patch ksvc {{ .CEAppName }} --type='merge' \
            -p="{\"spec\": {\"traffic\": [{\"percent\": 0, \"latestRevision\": true}, \
            {\"percent\": 100, \"revisionName\": \"$penultimateRevisionName\"}]}}"