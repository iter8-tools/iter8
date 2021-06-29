#!/bin/bash

set -e

export EXPERIMENT=quickstart-exp

# Ensure ITER8 environment variable is set
if [[ -z ${ITER8} ]]; then
    echo "ITER8 environment variable needs to be set to the root folder of Iter8"
    exit 1
else
    echo "ITER8 is set to " $ITER8
fi

# Check if experiment has completed
completed="Completed"
stage=$(kubectl get experiment ${EXPERIMENT} -o json | jq -r .status.stage)
if [[ $stage = $completed ]]; then
    echo "Experiment has Completed"
else
    echo "Experiment must be $completed;" 
    echo "Experiment is $stage;"
    exit 1
fi

# Check if versionRecommendedForPromotion is candidate
candidate="sample-app-v2"
vrfp=$(kubectl get experiment ${EXPERIMENT} -o json | jq -r .status.versionRecommendedForPromotion)
if [[ $vrfp = $candidate ]]; then
    echo "versionRecommendedForPromotion is $vrfp"
else
    echo "versionRecommendedForPromotion must be $candidate; is" $vrfp
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
    echo "percent is $percent"
else
    echo "percent must be $actualPercent; is" $percent
    exit 1
fi

set +e
