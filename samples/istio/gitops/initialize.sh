#!/bin/sh

rm -f experiment.yaml
rm -f productpage-candidate.yaml
rm -f fortio.yaml
git add -A ./; git commit -m "initialize"; git push origin head
