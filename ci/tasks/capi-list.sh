#!/bin/bash

set -eux

curl -s -H "Authorization: token ${API_TOKEN}" \
  https://api.github.com/repos/cloudfoundry/capi-release/releases \
  | jq -r '.[] | {tag_name,body,published_at}' \
  > releases

jq '.body' releases \
  | grep -F '**CC API Version:' \
  | grep -oE "2\.[0-9]+\.[0-9]+" \
  > api

jq -r '.published_at' releases > date

sed '/^,/d' <(paste -d,  api date)
