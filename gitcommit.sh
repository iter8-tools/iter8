#!/bin/bash
echo "gitCommit: $(git rev-parse HEAD)" > cmd/gitcommit.txt
