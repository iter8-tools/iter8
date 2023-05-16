#!/bin/sh
kubectl delete isvc wisdom-primary wisdom-candidate
kubectl delete deploy sleep
kubectl delete svc wisdom
kubectl delete vs wisdom
kubectl delete cm wisdom wisdom-input
