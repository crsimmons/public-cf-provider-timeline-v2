#!/bin/bash

set -eux

tag=$(cat capi-release/tag)

version=$(
  grep -F '**CC API Version:' capi-release/body \
  | grep -oE "2\.[0-9]+\.[0-9]+"
)

date=$(
  curl -s -H "Authorization: token ${API_TOKEN}" "https://api.github.com/repos/cloudfoundry/capi-release/releases/tags/${tag}" \
  | jq -r '.published_at'
)

echo "${version},${date}"
