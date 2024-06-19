#!/usr/bin/env bash

# Default values for optional arguments
: "${CERT_MANAGER_VERSION:=1.15.1}"
: "${JAEGER_VERSION:=1.57.0}"
: "${OTEL_CHART_VERSION:=0.97.1}"
: "${GRAFANA_CHART_VERSION:=8.3.2}"

# Bash settings
set -euxo pipefail

# Create the observability namespace
kubectl get namespace observability || kubectl create namespace observability

# Install cert-manager
helm upgrade --install cert-manager cert-manager --repo https://charts.jetstack.io \
  --namespace cert-manager --create-namespace \
  --description "Certificates manager for Ingress and internal traffic." \
  --values ./deploy/local/security/cert-manager/values.yaml \
  --version "v${CERT_MANAGER_VERSION}"

# Install Kubernetes Metrics Server
helm upgrade --install metrics-server metrics-server --repo https://kubernetes-sigs.github.io/metrics-server \
  --namespace observability \
  --description "Kubernetes Metrics server." \
  --set 'args={"--kubelet-insecure-tls"}'

# Install Prometheus
helm upgrade --install prometheus prometheus --repo https://prometheus-community.github.io/helm-charts \
  --namespace observability \
  --description "Prometheus metrics database." \
  --values ./deploy/local/observability/prometheus/prometheus-values.yaml

# Install the OTEL collector agent
helm upgrade --install otel-agent opentelemetry-collector --repo https://open-telemetry.github.io/opentelemetry-helm-charts \
  --namespace observability \
  --description "Agent (daemonset) OTEL collector for pull telemetry." \
  --values ./deploy/local/observability/otel/collector/agent-values.yaml \
  --version "${OTEL_CHART_VERSION}"
helm upgrade --install otel-gateway opentelemetry-collector --repo https://open-telemetry.github.io/opentelemetry-helm-charts \
  --namespace observability \
  --description "Gateway (deployment) OTEL collector for push telemetry." \
  --values ./deploy/local/observability/otel/collector/gateway-values.yaml \
  --version "${OTEL_CHART_VERSION}"

# Install Jaeger for traces
kubectl apply --namespace observability -f "https://github.com/jaegertracing/jaeger-operator/releases/download/v${JAEGER_VERSION}/jaeger-operator.yaml"

# Install Jaeger instance
until kubectl apply --namespace observability -f ./deploy/local/observability/jaeger/instance.yaml 2> /dev/null; do sleep 15; done

# Install Grafana
kubectl apply -k ./deploy/local/observability/grafana/config
helm upgrade --install grafana grafana --repo https://grafana.github.io/helm-charts \
  --namespace observability \
  --description "Grafana visualization hub." \
  --values ./deploy/local/observability/grafana/grafana-values.yaml \
  --version "${GRAFANA_CHART_VERSION}"
