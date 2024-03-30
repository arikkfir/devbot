#!/usr/bin/env bash

SNAPSHOTTER_BRANCH="release-7.0"
SNAPSHOTTER_VERSION="v7.0.1"

set -exuo pipefail

# Create a kind cluster
kind create cluster -n devbot

# Deploy the CRDs for the CSI driver
kubectl apply -f "https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/${SNAPSHOTTER_BRANCH}/client/config/crd/snapshot.storage.k8s.io_volumesnapshotclasses.yaml"
kubectl apply -f "https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/${SNAPSHOTTER_BRANCH}/client/config/crd/snapshot.storage.k8s.io_volumesnapshotcontents.yaml"
kubectl apply -f "https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/${SNAPSHOTTER_BRANCH}/client/config/crd/snapshot.storage.k8s.io_volumesnapshots.yaml"
kubectl apply -f "https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/${SNAPSHOTTER_VERSION}/deploy/kubernetes/snapshot-controller/rbac-snapshot-controller.yaml"
kubectl apply -f "https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/${SNAPSHOTTER_VERSION}/deploy/kubernetes/snapshot-controller/setup-snapshot-controller.yaml"

# Deploy the CSI driver from its Git repository
rm -rf /tmp/csi-driver-host-path && git clone https://github.com/kubernetes-csi/csi-driver-host-path.git /tmp/csi-driver-host-path
pushd /tmp/csi-driver-host-path
trap 'popd' EXIT
./deploy/kubernetes-latest/deploy.sh
kubectl rollout status statefulset/csi-hostpath-socat -n default
kubectl rollout status statefulset/csi-hostpathplugin -n default

# Verify CSI driver works
CURRENT_CONTEXT="$(kubectl config current-context)"
OLD_NAMESPACE="$(kubectl config view | yq ".contexts[]|select(.name==\"${CURRENT_CONTEXT}\")|.context.namespace // \"default\"")"
kubectl config set-context --current --namespace default
trap 'kubectl config set-context --current --namespace ${OLD_NAMESPACE}' EXIT
for i in ./examples/csi-storageclass.yaml ./examples/csi-pvc.yaml ./examples/csi-app.yaml; do kubectl apply -f $i; done
kubectl wait --for=condition=Ready -n default pod/my-csi-app --timeout=300s
kubectl get pv
kubectl get pvc
kubectl exec -it my-csi-app -- /bin/sh -c "touch /data/devbot-hello-world"
kubectl exec -it csi-hostpathplugin-0 -c hostpath -- /bin/sh -c "find / -name devbot-hello-world"
MATCHES=$(kubectl exec -it csi-hostpathplugin-0 -c hostpath -- /bin/sh -c "find / -name devbot-hello-world" | wc -l | tr -d '\t ')
[[ "${MATCHES}" == "2" ]] || (echo "Test file not found!" && exit 1)
for i in ./examples/csi-app.yaml ./examples/csi-pvc.yaml ./examples/csi-storageclass.yaml; do kubectl delete -f $i; done
