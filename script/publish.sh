#!/bin/bash
set -euo pipefail

PREVIOUS_TAG="$(git describe --abbrev=0 --tags --match '*-e7-*' HEAD)"

if [[ ! "$PREVIOUS_TAG" =~ ^([0-9]+\.[0-9]+\.[0-9]+)-e7-[0-9a-f]+$ ]]; then
  echo "Unexpected previous tag: $PREVIOUS_TAG" >&2
  exit 1
fi

PREVIOUS_TAG_VERSION="${BASH_REMATCH[1]}"
SOURCE_SHA="$(git rev-parse --short HEAD)"
NEW_TAG="$PREVIOUS_TAG_VERSION-e7-$SOURCE_SHA"
APT_CACHE_BUST="${APT_CACHE_BUST:-$(date -u +%Y-%m-%d)}"

IFS=. read -r -a SEMVER <<< "$PREVIOUS_TAG_VERSION"

PLATFORMS=(linux/amd64 linux/arm64/v8)
SUFFIXES=(amd64 arm64v8)
BASE_IMAGES=()
WORKSPACE_IMAGES=()
for i in "${!PLATFORMS[@]}"; do
  PLATFORM="${PLATFORMS[$i]}"
  SUFFIX="${SUFFIXES[$i]}"
  make base APT_CACHE_BUST="$APT_CACHE_BUST" DOCKER_REGISTRY=e7build DOCKER_PLATFORM="$PLATFORM"
  BASE_TAG="e7build/gotenberg:base-$SUFFIX"
  BASE_ID="$(docker image inspect --format '{{.Id}}' "$BASE_TAG")"
  BASE_IMAGES[i]="e7build/gotenberg:base-$SUFFIX-${BASE_ID#sha256:}"
  docker tag "$BASE_ID" "${BASE_IMAGES[$i]}"

  STABLE_TAG="e7build/gotenberg:stable-$SUFFIX"
  STABLE_ID="$(docker image inspect --format '{{.Id}}' "$STABLE_TAG")"
  STABLE_IMAGE="e7build/gotenberg:stable-$SUFFIX-${STABLE_ID#sha256:}"
  docker tag "$STABLE_ID" "$STABLE_IMAGE"
  make workspace PREPARE=0 STABLE_IMAGE="$STABLE_IMAGE" DOCKER_REGISTRY=e7build DOCKER_PLATFORM="$PLATFORM"
  WORKSPACE_TAG="e7build/gotenberg:workspace-$SUFFIX"
  WORKSPACE_ID="$(docker image inspect --format '{{.Id}}' "$WORKSPACE_TAG")"
  WORKSPACE_IMAGES[i]="e7build/gotenberg:workspace-$SUFFIX-${WORKSPACE_ID#sha256:}"
  docker tag "$WORKSPACE_ID" "${WORKSPACE_IMAGES[$i]}"
done

for i in "${!PLATFORMS[@]}"; do
  make image PREPARE=0 BASE_IMAGE="${BASE_IMAGES[$i]}" WORKSPACE_IMAGE="${WORKSPACE_IMAGES[$i]}" VERSION="$NEW_TAG" DOCKER_REGISTRY=e7build DOCKER_PLATFORM="${PLATFORMS[$i]}"
done

if [[ "$NEW_TAG" != "$PREVIOUS_TAG" ]]; then
  git tag "$NEW_TAG"
fi
git push origin master "$NEW_TAG"

docker push "e7build/gotenberg:$NEW_TAG-amd64"
docker push "e7build/gotenberg:$NEW_TAG-arm64v8"

for NEW_TAG_VARIANT in "$NEW_TAG" "${SEMVER[0]}.${SEMVER[1]}.${SEMVER[2]}" "${SEMVER[0]}.${SEMVER[1]}" "${SEMVER[0]}" latest; do
  docker buildx imagetools create \
    --tag "e7build/gotenberg:$NEW_TAG_VARIANT" \
    "e7build/gotenberg:$NEW_TAG-amd64" \
    "e7build/gotenberg:$NEW_TAG-arm64v8"
  echo "Published e7build/gotenberg:$NEW_TAG_VARIANT"
done
