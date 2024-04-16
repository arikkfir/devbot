#!/usr/bin/env bash

NS="$(k get ns -oname | grep -v kube | grep -v devbot | grep -v local-path-storage | grep -v default | cut -d/ -f2)"
[[ -z "${NS}" ]] && echo "No namespace to debug found" >&2 && exit 1

set -euo pipefail

RED='\033[0;31m'
NORMAL='\033[0m'

echo -e "${RED}=========[ NAMESPACE: ${NS} ]=========${NORMAL}"
echo -e
echo -e

echo -e "${RED}===[ devbot/v1/Repositories ]===${NORMAL}"
kubectl get -n "${NS}" repositories.devbot.kfirs.com --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ devbot/v1/Applications ]===${NORMAL}"
kubectl get -n "${NS}" applications.devbot.kfirs.com --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ devbot/v1/Environments ]===${NORMAL}"
kubectl get -n "${NS}" environments.devbot.kfirs.com --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ devbot/v1/Deployments ]===${NORMAL}"
kubectl get -n "${NS}" deployments.devbot.kfirs.com --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ batch/v1/Jobs ]===${NORMAL}"
kubectl get -n "${NS}" jobs.batch --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ apps/v1/Deployments ]===${NORMAL}"
kubectl get -n "${NS}" deployments.apps --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ v1/Pods ]===${NORMAL}"
kubectl get -n "${NS}" pods --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
for j in $(kubectl get -n "${NS}" jobs --sort-by='{.metadata.creationTimestamp}' -oname | cut -d'/' -f2); do
  echo -e "${RED}===[ job/${j} ]===${NORMAL}"
  kubectl logs -n "${NS}" jobs.batch/${j}
  echo -e
  echo -e
done
