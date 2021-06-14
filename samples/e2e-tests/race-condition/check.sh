#!/bin/bash

set -e

EXPERIMENT=race-condition

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

# Check if winner is baseline
expectedwinner="sample-app-v1"
winner=$(kubectl get experiment ${EXPERIMENT} -o json | jq -r .status.analysis.winnerAssessment.data.winner)
if [[ $winner = $expectedwinner ]]; then
    echo "winner is $winner"
else
    echo "winner must be $expectedwinner; is" $winner
    dump; exit 1
fi

set +e
