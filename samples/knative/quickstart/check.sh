#!/bin/bash

set -e

# Ensure ITER8 environment variable is set
if [[ -z ${ITER8} ]]; then
    echo "ITER8 environment variable needs to be set to the root folder of Iter8"
    exit 1
else
    echo "ITER8 is set to " $ITER8
fi

# Check if experiment has completed
completed="Completed"
stage=$(kubectl get experiment quickstart-exp -o json | jq -r .status.stage)
if [[ $stage = $completed ]]; then
    echo "Experiment has Completed"
else
    echo "Experiment must be $completed;" 
    echo "Experiment is $stage;"
    exit 1
fi

# Check if no packets have been lost by Fortio
pod_name=$(kubectl get pods --selector=job-name=fortio -o jsonpath='{.items[*].metadata.name}')
kubectl cp default/"$pod_name":shared/fortiooutput.json /tmp/fortiooutput.json -c busybox

REQUESTSTOTAL=$(jq -r .DurationHistogram.Count /tmp/fortiooutput.json)
REQUESTS200=$(jq -r '.RetCodes."200"' /tmp/fortiooutput.json)
if [[ $REQUESTSTOTAL -eq $REQUESTS200 ]]; then
    echo "Packets were not lost"
else
    echo "Packets were lost"
    echo "total requests:" $REQUESTSTOTAL
    echo "200 requests:" $REQUESTS200
    exit 1
fi

# Check if versionRecommendedForPromotion is candidate
candidate="candidate"
vrfp=$(kubectl get experiment quickstart-exp -o json | jq -r .status.versionRecommendedForPromotion)
if [[ $vrfp = $candidate ]]; then
    echo "versionRecommendedForPromotion is $vrfp"
else
    echo "versionRecommendedForPromotion must be candidate; is" $vrfp
    exit 1
fi

# Check if latest revision is true
latestRevision=true
lrStatus=$(kubectl get ksvc sample-app -o json | jq -r '.spec.traffic[0].latestRevision')
if [[ $lrStatus = $latestRevision ]]; then
    echo "latestRevision is true"
else
    echo "latestRevision must be true; is" $lrStatus
    exit 1
fi

#check if traffic percent is 100
percent=100
actualPercent=$(kubectl get ksvc sample-app -o json | jq -r '.spec.traffic[0].percent')
if [[ $actualPercent -eq $percent ]]; then
    echo "percent is 100"
else
    echo "percent must be 100; is" $percent
    exit 1
fi

set +e
