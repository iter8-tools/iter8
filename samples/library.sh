
# method to report failure
# expected usage in test cases: trap "reportFailure" EXIT
reportFailure () {
    status=$?
    if (( $status != 0 )); then
        echo -e "\033[0;31mFAILED:\033[0m $0"
    fi
    exit $status
}

# method to check the value of a field in the experiment
checkObjectField() {
    __name="${1}"        # print name
    __expected=${2}      # expected value of field
    __obj="${3}"         # Kubernetes object to be checked
    __path="${4}"        # jq expression to value of field in $__obj
    __ns="${5:-default}" # namespace (default is 'default')

    __actual=$(kubectl --namespace ${__ns} get ${__obj} -o json | jq -r ${__path})
    __message="Expecting ${__expected}; is ${__actual}"
    echo "Checking '${__name}' for ${__obj}; ${__message}"
    if [[ ${__actual} != ${__expected} ]]; then
        echo -e "\033[0;31mERROR:\033[0m ${__message}"
        exit 1
    fi
}

# Check that no packets have been lost by Fortio
# Assumes Fortio job in default namesapce
verifyNoPacketLoss() {
    pod_name=$(kubectl get pods --selector=job-name=fortio -o jsonpath='{.items[*].metadata.name}')
    kubectl cp default/"$pod_name":shared/fortiooutput.json /tmp/fortiooutput.json -c busybox

    REQUESTSTOTAL=$(jq -r .DurationHistogram.Count /tmp/fortiooutput.json)
    REQUESTS200=$(jq -r '.RetCodes."200"' /tmp/fortiooutput.json)
    if [[ $REQUESTSTOTAL -eq $REQUESTS200 ]]; then
        echo "Packets were not lost"
    else
        echo "\033[0;31mERROR:\033[0m Packets were lost"
        echo "total requests:" $REQUESTSTOTAL
        echo "200 requests:" $REQUESTS200
        exit 1
    fi
}
