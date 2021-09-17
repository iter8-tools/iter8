#!/usr/bin/env python

# Used by post-release-changes.yaml
# Change the install version in mkdocs.yaml

# Usage: python3 change-install-version.py <version number>

# Note: For some reason, ruamel.yaml does not correctly dump tags.
#
# For example, 
# !!python/name:materialx.emoji.twemoji
# is outputted as
# !%21python/name:materialx.emoji.twemoji
# from mkdocs/mkdocs.yaml
#
# Unsafe loading corrects this problem but has other issues, such as
# outputting tags in a strange format:
# !!python/name:materialx.emoji.twemoji ''
# as well as deleting all the comments.
#
# PyYAML, an alternative, will not preserve comments. Unsafe loading also
# outputs tags in the same strange format.

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