#!/bin/sh
kubectl delete deploy sleep
kubectl delete svc greatest-wisdom
kubectl delete vs greatest-wisdom
kubectl delete ns stable candidate
