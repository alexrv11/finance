terraform {
  required_version = ">= 1.6"

  required_providers {
    kind = {
      source  = "tehcyx/kind"
      version = "~> 0.4"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.31"
    }
  }
}

provider "kind" {}

provider "kubernetes" {
  host                   = kind_cluster.finance.endpoint
  cluster_ca_certificate = base64decode(kind_cluster.finance.cluster_ca_certificate)
  client_certificate     = base64decode(kind_cluster.finance.client_certificate)
  client_key             = base64decode(kind_cluster.finance.client_key)
}

# ─── kind cluster ────────────────────────────────────────────────────────────
# 1 control-plane + 2 worker nodes mirrors a minimal staging topology locally.

resource "kind_cluster" "finance" {
  name           = var.cluster_name
  wait_for_ready = true

  kind_config {
    kind        = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"

    node {
      role = "control-plane"

      # Expose NodePorts so services can be reached from localhost
      extra_port_mappings {
        container_port = 30080
        host_port      = 8080
        protocol       = "TCP"
      }
      extra_port_mappings {
        container_port = 30090
        host_port      = 9090
        protocol       = "TCP"
      }
    }

    node {
      role = "worker"
    }

    node {
      role = "worker"
    }
  }
}

# ─── Namespace ───────────────────────────────────────────────────────────────

resource "kubernetes_namespace" "finance" {
  metadata {
    name = "finance"

    labels = {
      environment = "local"
      managed-by  = "terraform"
    }
  }

  depends_on = [kind_cluster.finance]
}

# ─── ConfigMap — shared service discovery ────────────────────────────────────

resource "kubernetes_config_map" "service_config" {
  metadata {
    name      = "service-config"
    namespace = kubernetes_namespace.finance.metadata[0].name
  }

  data = {
    SEED_GRPC_ADDR      = "seed:50051"
    ANALYTICS_GRPC_ADDR = "analytics:50052"
    ALERT_GRPC_ADDR     = "alert:50053"
    ENVIRONMENT         = "local"
  }
}
