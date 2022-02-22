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

# set LOG_LEVEL for iter8 commands
export LOG_LEVEL="${INPUT_LOGLEVEL}"

echo "Fetch experiment"
$ITER8 hub -e ${INPUT_CHART}
cd $(basename ${INPUT_CHART})

echo "Identify values file"
OPTIONS=""
if [[ ! -z "${INPUT_VALUESFILE}" ]]; then
  OPTIONS="$OPTIONS -f values.yaml -f ../${INPUT_VALUESFILE}"
fi

echo "Create experiment.yaml for inspection"
echo "$ITER8 run --dry $OPTIONS"
$ITER8 run --dry $OPTIONS
cat experiment.yaml

echo "Run Experiment"
$ITER8 run $OPTIONS

echo "Log result"
$ITER8 report

echo "Run completed; verifying completeness"
# return 0 if satisfied; else non-zero
$ITER8 assert -c completed -c noFailure -c slos
