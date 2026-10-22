#!/bin/bash
set -euo pipefail

PREVIOUS_TAG="$(git describe --abbrev=0 --tags --match '*-e7-*' HEAD)"

if [[ ! "$PREVIOUS_TAG" =~ ^([0-9]+\.[0-9]+\.[0-9]+)-e7-[0-9a-f]+(-[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)?$ ]]; then
  echo "Unexpected previous tag: $PREVIOUS_TAG" >&2
  exit 1
fi

PREVIOUS_TAG_VERSION="${BASH_REMATCH[1]}"
SOURCE_SHA="$(git rev-parse --short HEAD)"
APT_CACHE_BUST="${APT_CACHE_BUST:-$(date -u +%Y-%m-%d)}"

IFS=. read -r -a SEMVER <<< "$PREVIOUS_TAG_VERSION"

PLATFORMS=(linux/amd64 linux/arm64/v8)
SUFFIXES=(amd64 arm64v8)
BASE_IMAGES=()
WORKSPACE_IMAGES=()
RELEASE_CHROMIUM_VERSION=

chromium_version() {
  local output
  output="$(docker run --rm --platform "$1" --entrypoint chromium "$2" --version)"
  if [[ ! "$output" =~ ^Chromium\ ([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)([[:space:]]|$) ]]; then
    echo "Unexpected Chromium version from $2: $output" >&2
    return 1
  fi
  printf '%s\n' "${BASH_REMATCH[1]}"
}

for i in "${!PLATFORMS[@]}"; do
  PLATFORM="${PLATFORMS[$i]}"
  SUFFIX="${SUFFIXES[$i]}"
  make base APT_CACHE_BUST="$APT_CACHE_BUST" DOCKER_REGISTRY=e7build DOCKER_PLATFORM="$PLATFORM"
  BASE_TAG="e7build/gotenberg:base-$SUFFIX"
  BASE_ID="$(docker image inspect --format '{{.Id}}' "$BASE_TAG")"
  BASE_IMAGES[i]="e7build/gotenberg:base-$SUFFIX-${BASE_ID#sha256:}"
  docker tag "$BASE_ID" "${BASE_IMAGES[$i]}"

  PLATFORM_CHROMIUM_VERSION="$(chromium_version "$PLATFORM" "${BASE_IMAGES[$i]}")"
  if [[ -n "$RELEASE_CHROMIUM_VERSION" && "$PLATFORM_CHROMIUM_VERSION" != "$RELEASE_CHROMIUM_VERSION" ]]; then
    echo "Chromium versions differ: $RELEASE_CHROMIUM_VERSION and $PLATFORM_CHROMIUM_VERSION ($PLATFORM)" >&2
    exit 1
  fi
  RELEASE_CHROMIUM_VERSION="$PLATFORM_CHROMIUM_VERSION"

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

NEW_TAG="$PREVIOUS_TAG_VERSION-e7-$SOURCE_SHA-$RELEASE_CHROMIUM_VERSION"
if git show-ref --verify --quiet "refs/tags/$NEW_TAG"; then
  echo "Release tag already exists: $NEW_TAG" >&2
  exit 1
fi

for i in "${!PLATFORMS[@]}"; do
  make image PREPARE=0 BASE_IMAGE="${BASE_IMAGES[$i]}" WORKSPACE_IMAGE="${WORKSPACE_IMAGES[$i]}" VERSION="$NEW_TAG" DOCKER_REGISTRY=e7build DOCKER_PLATFORM="${PLATFORMS[$i]}"
  FINAL_VERSION="$(chromium_version "${PLATFORMS[$i]}" "e7build/gotenberg:$NEW_TAG-${SUFFIXES[$i]}")"
  if [[ "$FINAL_VERSION" != "$RELEASE_CHROMIUM_VERSION" ]]; then
    echo "Final image Chromium version differs: expected $RELEASE_CHROMIUM_VERSION, got $FINAL_VERSION (${PLATFORMS[$i]})" >&2
    exit 1
  fi
done

git tag "$NEW_TAG"
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
