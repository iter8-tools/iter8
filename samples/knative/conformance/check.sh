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
stage=$(kubectl get experiment ${EXPERIMENT} -o json | jq -r .status.stage)
if [[ $stage = $completed ]]; then
    echo "Experiment has Completed"
else
    echo "Experiment must be $completed;" 
    echo "Experiment is $stage;"
    exit 1
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
