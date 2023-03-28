#!/bin/sh
kubectl delete ns primary candidate
kubectl delete deploy sleep
kubectl delete svc wisdom
kubectl delete vs wisdom wisdom-mirror
kubectl delete cm wisdom wisdom-input
