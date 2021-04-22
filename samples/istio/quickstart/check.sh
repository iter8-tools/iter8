#!/bin/bash

set -e

# dump logs from iter8 pods
dump() {
    # dump handler logs
    for pod in $(kubectl -n iter8-system get po --selector=iter8/experimentName=${EXPERIMENT} -o jsonpath='{.items[*].metadata.name}'); do 
        kubectl -n iter8-system logs $pod
    done

    # dump controller logs
    kubectl -n iter8-system logs $(kubectl -n iter8-system get po --selector=control-plane=controller-manager -o jsonpath='{.items[0].metadata.name}')
    
    # dump analytics logs
    kubectl -n iter8-system logs $(kubectl -n iter8-system get po --selector=app=iter8-analytics -o jsonpath='{.items[0].metadata.name}')
}

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
    dump; exit 1
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
    dump; exit 1
fi

# Check if versionRecommendedForPromotion is B
expectedVrfp="B"
vrfp=$(kubectl get experiment $EXPERIMENT -o json | jq -r .status.versionRecommendedForPromotion)
if [[ $vrfp = $expectedVrfp ]]; then
    echo "versionRecommendedForPromotion is $vrfp"
else
    echo "versionRecommendedForPromotion must be $expectedVrfp; is" $vrfp
    dump; exit 1
fi

# Check if latest revision is true
expectedSubset="productpage-v3"
subset=$(kubectl -n bookinfo-iter8 get vs bookinfo -o json | jq -r '.spec.http[0].route[0].destination.subset')
if [[ $subset = $expectedSubset ]]; then
    echo "subset is $subset"
else
    echo "subset must be $expectedSubset; is" $subset
    dump; exit 1
fi

#check if traffic percent is 100
percent=100
actualPercent=$(kubectl -n bookinfo-iter8 get vs bookinfo -o json | jq -r '.spec.http[0].route[0].weight')
if [[ $actualPercent -eq $percent ]]; then
    echo "percent is 100"
else
    echo "percent must be 100; is" $percent
    dump; exit 1
fi

set +e
