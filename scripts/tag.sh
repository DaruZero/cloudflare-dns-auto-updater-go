#!/bin/bash

# This script increments the version number based on the value of the
# environment variable TAG_INC. If TAG_INC is not set, the version's patch
# number is incremented by default. If TAG_INC is set to "major" or "minor",
# the major or minor number is incremented, respectively.

# Exit immediately if a command exits with a non-zero status.
set -e

# Get the current version number.
current_version=$(git describe --abbrev=0 --tags)

# Split the version number into its components, taking into account that the
# version number may be in the form "v1.2.3".
IFS='.' read -r -a version_parts <<< "${current_version#v}"

# Get the value of the TAG_INC environment variable.
tag_inc=${TAG_INC:-patch}

# Increment the appropriate version number.
case "$tag_inc" in
  major)
    ((version_parts[0]++))
    ;;
  minor)
    ((version_parts[1]++))
    ;;
  patch)
    ((version_parts[2]++))
    ;;
  *)
    echo "Unknown value for TAG_INC: $tag_inc"
    exit 1
    ;;
esac

# Join the version number components into a string.
new_version=$(IFS='.'; echo "${version_parts[*]}")

# Create a new tag.
git tag -s -m "v$new_version" "v$new_version"