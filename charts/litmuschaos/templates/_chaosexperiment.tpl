{{- define "chaosexperiment" -}}
apiVersion: litmuschaos.io/v1alpha1
description:
  message: |
    Deletes a pod belonging to a deployment/statefulset/daemonset
kind: ChaosExperiment
metadata:
  name: pod-delete
  labels:
    app.kubernetes.io/part-of: litmus
    app.kubernetes.io/component: chaosexperiment
    app.kubernetes.io/version: {{ .Chart.AppVersion }}
spec:
  definition:
    scope: Namespaced
    permissions:
      - apiGroups:
          - ""
          - "apps"
          - "apps.openshift.io"
          - "argoproj.io"
          - "batch"
          - "litmuschaos.io"
        resources:
          - "deployments"
          - "jobs"
          - "pods"
          - "pods/log"
          - "replicationcontrollers"
          - "deployments"
          - "statefulsets"
          - "daemonsets"
          - "replicasets"
          - "deploymentconfigs"
          - "rollouts"
          - "pods/exec"
          - "events"
          - "chaosengines"
          - "chaosexperiments"
          - "chaosresults"
        verbs:
          - "create"
          - "list"
          - "get"
          - "patch"
          - "update"
          - "delete"
          - "deletecollection"
    image: "litmuschaos/go-runner:{{ .Chart.AppVersion }}"
    imagePullPolicy: Always
    args:
    - -c
    - ./experiments -name pod-delete
    command:
    - /bin/bash
    env:

    - name: TOTAL_CHAOS_DURATION
      value: '15'

    # Period to wait before and after injection of chaos in sec
    - name: RAMP_TIME
      value: ''

    - name: FORCE
      value: 'true'

    - name: CHAOS_INTERVAL
      value: '5'

    ## percentage of total pods to target
    - name: PODS_AFFECTED_PERC
      value: ''

    - name: LIB
      value: 'litmus'    

    - name: TARGET_PODS
      value: ''

    ## it defines the sequence of chaos execution for multiple target pods
    ## supported values: serial, parallel
    - name: SEQUENCE
      value: 'parallel'
      
    labels:
      app.kubernetes.io/part-of: litmus
      app.kubernetes.io/component: experiment-job
      app.kubernetes.io/version: {{ .Chart.AppVersion }}
{{- end }}
