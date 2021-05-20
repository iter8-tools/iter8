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
REQUESTSTOTAL=$(jq -r .DurationHistogram.Count /tmp/fortiooutput1.json)
REQUESTS200=$(jq -r '.RetCodes."200"' /tmp/fortiooutput1.json)
if [[ $REQUESTSTOTAL -eq $REQUESTS200 ]]; then
    echo "Packets were not lost in fortiooutput1"
else
    echo "Packets were lost in fortiooutput1"
    echo "total requests:" $REQUESTSTOTAL
    echo "200 requests:" $REQUESTS200
    exit 1
fi


# Check if no packets have been lost by Fortio
REQUESTSTOTAL=$(jq -r .DurationHistogram.Count /tmp/fortiooutput2.json)
REQUESTS200=$(jq -r '.RetCodes."200"' /tmp/fortiooutput2.json)
if [[ $REQUESTSTOTAL -eq $REQUESTS200 ]]; then
    echo "Packets were not lost in fortiooutput2"
else
    echo "Packets were lost in fortiooutput2"
    echo "total requests:" $REQUESTSTOTAL
    echo "200 requests:" $REQUESTS200
    exit 1
fi

# Check if versionRecommendedForPromotion is flowers-v2
expectedVrfp="flowers-v2"
vrfp=$(kubectl get experiment $EXPERIMENT -o json | jq -r .status.versionRecommendedForPromotion)
if [[ $vrfp = $expectedVrfp ]]; then
    echo "versionRecommendedForPromotion is $vrfp"
else
    echo "versionRecommendedForPromotion must be $expectedVrfp; is" $vrfp
    exit 1
fi

set +e
