#!/bin/sh

# give fortio deployment a random name so it restarts on a new experiment
RANDOM=`od -An -N4 -i /dev/random`
sed "s|  name: fortio-|  name: fortio-$RANDOM|" templates/fortio.yaml > ./fortio.yaml

# give experiment a random name so CI triggers new experiment each time a new app version is available
sed "s|name: gitops-exp|name: gitops-exp-$RANDOM|" templates/experiment.yaml > ./experiment.yaml

# use a random color for a new experiment candidate
declare -a colors=("red" "orange" "blue" "green" "yellow" "violet" "brown")
color=`expr $RANDOM % ${#colors[@]}`
sed "s|value: COLOR|value: \"${colors[$color]}\"|" templates/productpage-candidate.yaml > ./productpage-candidate.yaml

