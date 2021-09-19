#!/usr/bin/env python

# Used by post-release-changes.yaml
# Verify that version number follows semantic versioning

# Usage: python3 semver-check.py <version number>

import sys
import semantic_version

if len(sys.argv) < 2:
  raise ValueError('Please provide an install version as an argument.')

# Will throw error if the version number does not follow semantic versioning
semantic_version.Version(sys.argv[1])