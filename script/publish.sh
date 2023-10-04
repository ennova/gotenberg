#!/bin/bash
set -euo pipefail

PREVIOUS_TAG="$(git describe --abbrev=0 --tags HEAD)"

if ! echo "$PREVIOUS_TAG" | egrep -q "^[0-9.]+-e7-[0-9a-f]+$"; then
  echo "Unexpected previous tag: $PREVIOUS_TAG" >&2
  exit 1
fi

PREVIOUS_TAG_VERSION="$(echo "$PREVIOUS_TAG" | egrep -o "^[0-9.]+")"

NEW_TAG="$PREVIOUS_TAG_VERSION-e7-$(git rev-parse --short HEAD)"

SEMVER=( ${PREVIOUS_TAG_VERSION//./ } )
VERSION_LENGTH=${#SEMVER[@]}

if [ $VERSION_LENGTH -ne 3 ]; then
  echo "$VERSION is not semver."
  exit 1
fi

make image VERSION="$NEW_TAG" DOCKER_REGISTRY=e7build DOCKER_PLATFORM=linux/amd64
make image VERSION="$NEW_TAG" DOCKER_REGISTRY=e7build DOCKER_PLATFORM=linux/arm64/v8

if [[ "$NEW_TAG" != "$PREVIOUS_TAG" ]]; then
   git tag "$NEW_TAG"
fi

git push origin master "$NEW_TAG"

docker push e7build/gotenberg:$NEW_TAG-amd64
docker push e7build/gotenberg:$NEW_TAG-arm64v8

for NEW_TAG_VARIANT in $NEW_TAG ${SEMVER[0]}.${SEMVER[1]}.${SEMVER[2]} ${SEMVER[0]}.${SEMVER[1]} ${SEMVER[0]} latest; do
  docker manifest rm e7build/gotenberg:$NEW_TAG_VARIANT 2> /dev/null || true
  docker manifest create e7build/gotenberg:$NEW_TAG_VARIANT e7build/gotenberg:$NEW_TAG-{amd64,arm64v8}
  docker manifest push e7build/gotenberg:$NEW_TAG_VARIANT
done
