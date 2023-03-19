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
