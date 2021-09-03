#!/usr/bin/env python

# Used by ind.yaml
# Change the install version in mkdocs.yaml

# Usage: python3 change-install-version.py <version number>

import sys
from ruamel.yaml import YAML

if len(sys.argv) < 2:
  raise ValueError('Please provide an install version as an argument.')

new_install_version=sys.argv[1]
mkdocs_file_path='./mkdocs/mkdocs.yml'

with open(mkdocs_file_path, 'r') as fp:
  # Read YAML
  yaml=YAML()
  data=yaml.load(fp)

  # Replace install version
  data['extra']['iter8']['install_version']=new_install_version

with open(mkdocs_file_path, 'w') as fp:
  # Rewrite YAML
  yaml.dump(data, fp)