#!/bin/bash
set -euo pipefail

DEPLOYMENT="${DEPLOYMENT:-higress-console}"
NAMESPACE="${NAMESPACE:-higress-system}"
TAG="${1:-}"

usage() {
  echo "Usage: NAMESPACE=<ns> $0 <new-tag>"
  echo "Example: NAMESPACE=higress $0 v1.0.1"
  exit 1
}

if [ -z "$TAG" ]; then
  usage
fi

NS_FLAG=""
if [ -n "$NAMESPACE" ]; then
  NS_FLAG="-n $NAMESPACE"
fi

# 取当前镜像地址并替换 tag
CURRENT_IMAGE=$(kubectl get deployment "$DEPLOYMENT" $NS_FLAG \
  -o jsonpath='{.spec.template.spec.containers[0].image}')
NEW_IMAGE="${CURRENT_IMAGE%:*}:${TAG}"

echo "Patching deployment '$DEPLOYMENT'"
echo "  Old image: $CURRENT_IMAGE"
echo "  New image: $NEW_IMAGE"

kubectl set image deployment/"$DEPLOYMENT" "$DEPLOYMENT"="$NEW_IMAGE" $NS_FLAG

echo "Waiting for rollout to finish..."
kubectl rollout status deployment/"$DEPLOYMENT" $NS_FLAG
