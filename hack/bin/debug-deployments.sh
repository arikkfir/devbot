#!/usr/bin/env bash

set -euo pipefail

RED='\033[0;31m'
NORMAL='\033[0m'
NS="$(k get ns -oname | grep -v kube | grep -v devbot | grep -v local-path-storage | grep -v default | cut -d/ -f2)"

echo -e "${RED}=========[ NAMESPACE: ${NS} ]=========${NORMAL}"
echo -e
echo -e

echo -e "${RED}===[ APPLICATIONS ]===${NORMAL}"
kubectl get -n "${NS}" applications --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ ENVIRONMENTS ]===${NORMAL}"
kubectl get -n "${NS}" environments --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ DEPLOYMENTS ]===${NORMAL}"
kubectl get -n "${NS}" deployments.devbot.kfirs.com --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ JOBS ]===${NORMAL}"
kubectl get -n "${NS}" jobs --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
echo -e "${RED}===[ PODS ]===${NORMAL}"
kubectl get -n "${NS}" pods --sort-by='{.metadata.creationTimestamp}' -owide
echo -e
echo -e
for j in $(kubectl get -n "${NS}" jobs --sort-by='{.metadata.creationTimestamp}' -oname | cut -d'/' -f2); do
  echo -e "${RED}===[ job/${j} ]===${NORMAL}"
  kubectl logs -n "${NS}" jobs/${j}
  echo -e
  echo -e
done
