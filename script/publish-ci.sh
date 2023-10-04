#!/bin/bash
set -euo pipefail

aws ecr get-login-password --profile ennova-buildkite | skopeo login --username AWS --password-stdin 252974019764.dkr.ecr.ap-southeast-2.amazonaws.com/gotenberg

curl -sSf "https://hub.docker.com/v2/namespaces/e7build/repositories/gotenberg/tags?page_size=10" | jq '.results[].name' -r | while read TAG; do
  skopeo copy --all docker://e7build/gotenberg:$TAG docker://252974019764.dkr.ecr.ap-southeast-2.amazonaws.com/gotenberg:$TAG
done
