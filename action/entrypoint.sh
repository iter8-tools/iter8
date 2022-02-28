#!/bin/bash -l

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ITER8="/bin/iter8"

echo "Creating working directory"

WORK_DIR=`mktemp -d -p  "$DIR"`
if [[ ! "$WORK_DIR" || ! -d  "$WORK_DIR" ]]; then
  echo "Cound not create temporary working directory"
  exit 1
fi

# no need to cleanup

echo "Identify loglevel if set"
LOGLEVEL=""
if [[ ! -z "${INPUT_LOGLEVEL}" ]]; then
  LOGLEVEL="$LOGLEVEL -l ${INPUT_LOGLEVEL}"
fi

echo "Identify values file"
OPTIONS=""
if [[ ! -z "${INPUT_VALUESFILE}" ]]; then
  OPTIONS="$OPTIONS -f ${INPUT_VALUESFILE}"
fi

echo "Identify any chartVersionConstraint file"
if [[ ! -z "${INPUT_CHARTVERSION}" ]]; then
  OPTIONS="$OPTIONS -v ${INPUT_CHARTVERSION}"
fi

echo "Calling: $ITER8 launch -c ${INPUT_CHART} ${OPTIONS} ${LOGLEVEL}"
$ITER8 launch -c ${INPUT_CHART} ${OPTIONS} ${LOGLEVEL}

echo "Log benchmarks"
$ITER8 report ${LOGLEVEL}

echo "Experiment completed"
# return 0 if satisfied; else non-zero
if [[ "${INPUT_VALIDATESLOS}" == "true" ]]; then
  echo "Asserting SLOs satisfied"
  $ITER8 assert -c completed -c noFailure -c slos ${LOGLEVEL}
fi
