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
stage=$(kubectl get experiment $EXPERIMENT -o json | jq -r .status.stage)
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

# Check if versionRecommendedForPromotion is B
expectedVrfp="productpage-v2"
vrfp=$(kubectl get experiment $EXPERIMENT -o json | jq -r .status.versionRecommendedForPromotion)
if [[ $vrfp = $expectedVrfp ]]; then
    echo "versionRecommendedForPromotion is $vrfp"
else
    echo "versionRecommendedForPromotion must be $expectedVrfp; is" $vrfp
    exit 1
fi

# This experiment does not include a finish action; these are commented out.

# # Check if traffic percent to productpage-v2 is > 50
# expectedPercent=100
# actualPercent=$(kubectl -n bookinfo-iter8 get vs bookinfo -o json | jq -r '.spec.http[0].route[1].weight')
# if (( $actualPercent >= $expectedPercent )); then
#     echo "percent is $actualPercent"
# else
#     echo "percent must be greater than or equal to $expectedPercent; is" $actualPercent
#     exit 1
# fi

#check if traffic percent is 100
expectedPercent=100
actualPercent=$(kubectl -n bookinfo-iter8 get vs bookinfo -o json | jq -r '.spec.http[0].route[0].weight')
if [[ $actualPercent -eq $expectedPercent ]]; then
    echo "percent is 100"
else
    echo "percent must be $expectedPercent; is" $actualPercent
    exit 1
fi

set +e
