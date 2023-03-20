# initialize primary v1
./initialize.sh
# query
./sleep.sh
# in a new terminal
./execintosleep.sh
# inside the sleep pod
cd wisdom
source query.sh

# candidate v2
./v2-candidate.sh

# bump up traffic for candidate v2
./bumpweights.sh

# promote v2
./promote-v2.sh
kubectl delete isvc wisdom-candidate

# candidate v3
./v3-candidate.sh

# delete v3
kubectl delete isvc wisdom-candidate