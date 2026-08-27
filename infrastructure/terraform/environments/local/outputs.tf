output "cluster_name" {
  description = "kind cluster name"
  value       = kind_cluster.finance.name
}

output "cluster_endpoint" {
  description = "Kubernetes API server endpoint"
  value       = kind_cluster.finance.endpoint
}

output "namespace" {
  description = "Kubernetes namespace for all services"
  value       = kubernetes_namespace.finance.metadata[0].name
}

output "kubeconfig_command" {
  description = "Run this to point kubectl at the local cluster"
  value       = "kind get kubeconfig --name ${var.cluster_name} > ~/.kube/config-finance-local && export KUBECONFIG=~/.kube/config-finance-local"
}
