#!/bin/bash
set -e

readonly version="$1"

git tag $version &
git push origin $version &
goreleaser --rm-dist