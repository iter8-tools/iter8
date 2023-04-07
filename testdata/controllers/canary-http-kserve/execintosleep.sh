#!/bin/sh
# First, get $SLEEP_POD
SLEEP_POD=$(kubectl get pod --sort-by={metadata.creationTimestamp} -l app=sleep -o jsonpath={.items..metadata.name} | rev | cut -d' ' -f 1 | rev)
# Second, exec into it
kubectl exec --stdin --tty "${SLEEP_POD}" -c sleep -- /bin/sh
# Third, cd wisdom && source query.sh in order to query wisdom