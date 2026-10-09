#!/bin/bash

set -euo pipefail

go build -o /dev/null cmd/gotenberg/main.go
make image DOCKER_REGISTRY=e7build

clear

docker run \
  --rm -it \
  -p 3090:3000 \
  -m 256M \
  -e LOG_LEVEL=DEBUG \
  -e GOOGLE_CHROME_MAX_CONNECTIONS=2 \
  -e MAXIMUM_WEBHOOK_URL_TIMEOUT=120 \
  -e MAXIMUM_WAIT_TIMEOUT=900 \
  e7build/gotenberg:snapshot
