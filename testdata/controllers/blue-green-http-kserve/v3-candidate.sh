cat <<EOF | kubectl apply -f -
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: wisdom-candidate
  labels:
    app.kubernetes.io/name: wisdom
    app.kubernetes.io/version: v3
    iter8.tools/watch: "true"
spec:
  predictor:
    minReplicas: 1
    model:
      modelFormat:
        name: sklearn
      runtime: kserve-mlserver
      storageUri: "gs://seldon-models/sklearn/mms/lr_model"
EOF