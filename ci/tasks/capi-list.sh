#!/bin/bash

set -eux

releases=$(
  curl -s -H "Authorization: token ${API_TOKEN}" \
  https://api.github.com/repos/cloudfoundry/capi-release/releases \
  | jq -r '.[] | {tag_name,body,published_at}'
)

api=$(
  jq '.body' "${releases}" \
  | grep -F '**CC API Version:' \
  | grep -oE "2\.[0-9]+\.[0-9]+"
)

date=$(jq -r '.published_at' "${releases}")

paste -d,  "${api}" "${date}" | sed '/^,/d'
