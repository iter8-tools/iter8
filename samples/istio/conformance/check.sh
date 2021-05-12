#!/bin/bash

set -e -x

# Ensure ITER8 environment variable is set
if [[ -z ${ITER8} ]]; then
    echo "ITER8 environment variable needs to be set to the root folder of Iter8"
    exit 1
else
    echo "ITER8 is set to " $ITER8
fi

# Check if experiment has completed
checkObjectField "Condition Completed" "Completed" "experiment/${EXPERIMENT}" ".status.stage"

# Check that no packets have been lost
verifyNoPacketLoss

# Check if winner is found and is productpage-v1
checkObjectField "winnerFound" "true" "experiment/${EXPERIMENT}" ".status.analysis.winnerAssessment.data.winnerFound"
checkObjectField "winner" "productpage-v1" "experiment/${EXPERIMENT}" ".status.analysis.winnerAssessment.data.winner"

set +e +x
