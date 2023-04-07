#!/bin/sh
cat <<EOF | istioctl kube-inject -f - | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sleep
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sleep
  template:
    metadata:
      labels:
        app: sleep
    spec:
      containers:
      - name: sleep
        image: curlimages/curl
        command: ["/bin/sh", "-c", "sleep 3650d"]
        imagePullPolicy: IfNotPresent
        volumeMounts:
        - name: config-volume
          mountPath: /wisdom
      volumes:
      - name: config-volume
        configMap:
          name: wisdom-input
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: wisdom-input
data:
  input.json: |
    {
      "inputs": [
        {
          "name": "input-0",
          "shape": [2, 4],
          "datatype": "FP32",
          "data": [
            [6.8, 2.8, 4.8, 1.4],
            [6.0, 3.4, 4.5, 1.6]
          ]
        }
      ]
    }
  query.sh: |
    echo "curl -H 'Content-Type: application/json' http://wisdom.default/enlightenme -d @input.json"
    curl -H 'Content-Type: application/json' http://wisdom.default/enlightenme -d @input.json
EOF
