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
checkObjectField "Condition Completed" "Completed" "experiment/${EXPERIMENT}" ".status.stage"

# Check that no packets have been lost
verifyNoPacketLoss

# Check if versionRecommendedForPromotion is productpage-v2
checkObjectField "versionRecommendedForPromotion" "productpage-v2" "experiment/${EXPERIMENT}" ".status.versionRecommendedForPromotion"

# Check if winner is found and is productpage-v2
checkObjectField "winnerFound" "true" "experiment/${EXPERIMENT}" ".status.analysis.winnerAssessment.data.winnerFound"
checkObjectField "winner" "productpage-v2" "experiment/${EXPERIMENT}" ".status.analysis.winnerAssessment.data.winner"

# Check if VirtualService updated
checkObjectField "subset" "productpage-v2" "vs/bookinfo" ".spec.http[0].route[0].destination.subset" bookinfo-iter8

# Check if traffic percent is 100
checkObjectField "Traffic (percentage) to winner" 100 "vs/bookinfo" ".spec.http[0].route[0].weight" bookinfo-iter8

set +e
