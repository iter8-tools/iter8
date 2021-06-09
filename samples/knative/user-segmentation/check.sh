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
kubectl get experiment request-routing -o yaml
completed="Completed"
stage=$(kubectl get experiment request-routing -o json | jq -r .status.stage)
if [[ $stage = $completed ]]; then
    echo "Experiment has Completed"
else
    echo "Experiment must be $completed;" 
    echo "Experiment is $stage;"
    exit 1
fi

# Check if versionRecommendedForPromotion is candidate
candidate="sample-app-v2-green"
vrfp=$(kubectl get experiment request-routing -o json | jq -r .status.versionRecommendedForPromotion)
if [[ $vrfp = $candidate ]]; then
    echo "versionRecommendedForPromotion is $vrfp"
else
    echo "versionRecommendedForPromotion must be $candidate; is" $vrfp
    exit 1
fi

set +e
