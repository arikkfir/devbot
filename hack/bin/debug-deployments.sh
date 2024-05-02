#!/usr/bin/env bash

set -euo pipefail

RED='\033[0;31m'
NORMAL='\033[0m'

echo -e "${RED}===[ devbot/v1/Repositories ]===${NORMAL}"
kubectl get -A repositories.devbot.kfirs.com --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ devbot/v1/Applications ]===${NORMAL}"
kubectl get -A applications.devbot.kfirs.com --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ devbot/v1/Environments ]===${NORMAL}"
kubectl get -A environments.devbot.kfirs.com --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ devbot/v1/Deployments ]===${NORMAL}"
kubectl get -A deployments.devbot.kfirs.com --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ batch/v1/Jobs ]===${NORMAL}"
kubectl get -A jobs.batch --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ apps/v1/Deployments ]===${NORMAL}"
kubectl get -A deployments.apps --sort-by='{.metadata.creationTimestamp}' | grep -v kube-system | grep -v local-path-storage
echo -e
echo -e
echo -e "${RED}===[ v1/Pods ]===${NORMAL}"
kubectl get -A pods --sort-by='{.metadata.creationTimestamp}' -owide | grep -v kube-system | grep -v local-path-storage
echo -e
echo -e
for j in $(kubectl get -A jobs --sort-by='{.metadata.creationTimestamp}' -o jsonpath='{range .items[*]}{.metadata.namespace}/{.metadata.name} {end}'); do
  echo -e "${RED}===[ jobs.batch/${j} ]===${NORMAL}"
  NS=$(echo "${j}" | cut -d/ -f1)
  NAME=$(echo "${j}" | cut -d/ -f2)
  kubectl logs "jobs.batch/${NAME}" -n "${NS}"
  echo -e
  echo -e
done
