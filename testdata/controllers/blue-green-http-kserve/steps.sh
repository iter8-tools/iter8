# Start iter8 controller
POD_NAME=blue-leader-0 CONFIG_FILE=testdata/controllers/config.yaml go run main.go controllers -l trace
# initialize primary v1
./initialize.sh
# query
./sleep.sh
# in a new terminal
./execintosleep.sh
# inside the sleep pod
cd wisdom
source query.sh

# Explore what the initialization step entailed ... templated virtual service
less initialize.sh

# Explore status of virtual service ... 
kubectl get vs wisdom -o yaml

# check back on status of query ... 

# candidate v2
./v2-candidate.sh

source query.sh

# Explore what candidate release entails... 
less v2-candidate.sh

# Explore status of virtual service ... 
kubectl get vs wisdom -o yaml

# check back on status of warm up 

# bump up traffic for candidate v2
./bumpweights.sh

# Explore status of virtual service ... 
kubectl get vs wisdom -o yaml

# promote v2
./promote-v2.sh
