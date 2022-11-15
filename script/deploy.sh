#!/bin/bash
set -ueo pipefail

HEROKU_APP="${1?}"

set -x

docker pull --platform linux/amd64 e7build/gotenberg:latest
docker tag e7build/gotenberg:latest "registry.heroku.com/$HEROKU_APP/web:latest"
docker push "registry.heroku.com/$HEROKU_APP/web:latest"
heroku container:release web -a "$HEROKU_APP"
