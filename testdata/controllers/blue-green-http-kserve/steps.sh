# Start iter8 controller
PODNAME=blue-leader-0 CONFIG_FILE=testdata/controllers/config.yaml go run main.go controllers -l trace
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
