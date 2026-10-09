#!/bin/bash
set -euo pipefail

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
  echo "Usage: $0 <qa|prod> [tag]" >&2
  exit 2
fi

case "$1" in
  qa|prod) ENVIRONMENT="$1" ;;
  *)
    echo "Expected environment: qa or prod" >&2
    exit 2
    ;;
esac

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
if [ "$#" -eq 2 ]; then
  TAG="$2"
else
  TAG="$(git -C "$REPO_ROOT" describe --abbrev=0 --tags --match '*-e7-*' HEAD)"
fi

if [[ ! "$TAG" =~ ^[0-9]+\.[0-9]+\.[0-9]+-e7-[0-9a-f]+$ ]]; then
  echo "Unexpected release tag: $TAG" >&2
  exit 1
fi

SUBSCRIPTION="SYNCHRO Perform"
WEBAPP="$ENVIRONMENT-synchroperform-eus2-gotenberg"
RESOURCE_GROUP="$WEBAPP-rg"
REGISTRY="${ENVIRONMENT}synchroperformeus2"
REGISTRY_SERVER="$(az acr show \
  --subscription "$SUBSCRIPTION" \
  --name "$REGISTRY" \
  --query loginServer \
  --output tsv)"

if [ -z "$REGISTRY_SERVER" ]; then
  echo "No login server returned for registry: $REGISTRY" >&2
  exit 1
fi

SOURCE_IMAGE="e7build/gotenberg:$TAG-amd64"
TARGET_IMAGE="$REGISTRY_SERVER/gotenberg:$TAG"

echo "Deploying $SOURCE_IMAGE to $WEBAPP as $TARGET_IMAGE"
docker pull --platform linux/amd64 "$SOURCE_IMAGE"
az acr login --subscription "$SUBSCRIPTION" --name "$REGISTRY" --output none
docker tag "$SOURCE_IMAGE" "$TARGET_IMAGE"
docker push "$TARGET_IMAGE"

az webapp config container set \
  --subscription "$SUBSCRIPTION" \
  --resource-group "$RESOURCE_GROUP" \
  --name "$WEBAPP" \
  --container-image-name "$TARGET_IMAGE" \
  --container-registry-url "https://$REGISTRY_SERVER" \
  --output none

az webapp restart \
  --subscription "$SUBSCRIPTION" \
  --resource-group "$RESOURCE_GROUP" \
  --name "$WEBAPP"

echo "Restart requested for https://$WEBAPP.azurewebsites.net using $TARGET_IMAGE"
