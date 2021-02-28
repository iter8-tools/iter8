---
template: overrides/main.html
---

# Request Routing

## 1. Create app versions
```shell
kubectl apply -f $ITER8/samples/knative/requestrouting/services.yaml
```

## 2. Create Istio virtual service
```shell
kubectl apply -f $ITER8/samples/knative/requestrouting/routing-rule.yaml
```

## 3. Generate traffic
```shell
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.8.1 sh -
istio-1.8.1/bin/istioctl kube-inject -f $ITER8/samples/knative/requestrouting/curl.yaml | kubectl create -f -
cd $ITER8
```

## 4. Create experiment
```shell
kubectl wait --for=condition=Ready ksvc/sample-app-v1
kubectl wait --for=condition=Ready ksvc/sample-app-v2
kubectl apply -f $ITER8/samples/knative/requestrouting/experiment.yaml
```

## 5. Observe experiment
You can observe the experiment in realtime. Follow instructions in the three tabs below in three separate terminals.

=== "iter8ctl"
    ```shell
    while clear; do
    kubectl get experiment request-routing -o yaml | iter8ctl describe -f -
    sleep 4
    done
    ```

=== "kubectl get experiment"

    ```shell
    kubectl get experiment request-routing --watch
    ```

=== "kubectl get vs"
    ```shell
    kubectl get vs  routing-for-wakanda -o json | jq .spec.http[0].route
    ```


## 6. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/requestrouting/experiment.yaml
kubectl delete -f $ITER8/samples/knative/requestrouting/curl.yaml
kubectl delete -f $ITER8/samples/knative/requestrouting/routing-rule.yaml
kubectl delete -f $ITER8/samples/knative/requestrouting/services.yaml
```

