#!/usr/bin/env bash
# Deploy the full observability stack to the current kubectl context.
# Prerequisites: helm >= 3.12, kubectl configured and pointed at target cluster.

set -euo pipefail

NAMESPACE=monitoring

echo "→ Adding Helm repos..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

echo "→ Creating namespace..."
kubectl apply -f namespace.yaml

echo "→ Installing kube-prometheus-stack (Prometheus + Grafana + AlertManager)..."
helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace "$NAMESPACE" \
  --version 65.1.1 \
  --values prometheus/values.yaml \
  --wait

echo "→ Installing Tempo (distributed tracing)..."
helm upgrade --install tempo grafana/tempo \
  --namespace "$NAMESPACE" \
  --version 1.10.1 \
  --values tempo/values.yaml \
  --wait

echo "→ Installing Loki (log aggregation)..."
helm upgrade --install loki grafana/loki \
  --namespace "$NAMESPACE" \
  --version 6.16.0 \
  --values loki/values.yaml \
  --wait

echo "→ Installing Promtail (log shipper)..."
helm upgrade --install promtail grafana/promtail \
  --namespace "$NAMESPACE" \
  --version 6.16.6 \
  --values promtail/values.yaml \
  --wait

echo ""
echo "✓ Observability stack deployed."
echo ""
echo "  Grafana:"
echo "    kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80"
echo "    open http://localhost:3000  (admin / changeme)"
echo ""
echo "  Prometheus:"
echo "    kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090"
echo ""
echo "  AlertManager:"
echo "    kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-alertmanager 9093:9093"
