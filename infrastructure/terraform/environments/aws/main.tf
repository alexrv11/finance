terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.31"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = local.common_tags
  }
}

# Kubernetes provider authenticates against EKS using a short-lived token.
data "aws_eks_cluster_auth" "this" {
  name = module.eks.cluster_name
}

provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_ca_certificate)
  token                  = data.aws_eks_cluster_auth.this.token
}

# ─── Locals ──────────────────────────────────────────────────────────────────

locals {
  common_tags = {
    project     = "finance"
    environment = var.environment
    managed-by  = "terraform"
  }
}

# ─── VPC ─────────────────────────────────────────────────────────────────────

module "vpc" {
  source = "../../modules/vpc"

  name            = "finance-${var.environment}"
  cidr            = var.vpc_cidr
  azs             = var.azs
  public_subnets  = var.public_subnets
  private_subnets = var.private_subnets
  tags            = local.common_tags
}

# ─── EKS ─────────────────────────────────────────────────────────────────────

module "eks" {
  source = "../../modules/eks"

  cluster_name        = "finance-${var.environment}"
  kubernetes_version  = var.kubernetes_version
  vpc_id              = module.vpc.vpc_id
  subnet_ids          = module.vpc.private_subnet_ids
  node_instance_types = var.node_instance_types
  node_desired_size   = var.node_desired_size
  node_min_size       = var.node_min_size
  node_max_size       = var.node_max_size
  tags                = local.common_tags
}

# ─── ECR ─────────────────────────────────────────────────────────────────────

module "ecr" {
  source = "../../modules/ecr"

  project  = "finance"
  services = var.services
  tags     = local.common_tags
}

# ─── Kubernetes baseline ─────────────────────────────────────────────────────

resource "kubernetes_namespace" "finance" {
  metadata {
    name = "finance"

    labels = {
      environment = var.environment
      managed-by  = "terraform"
    }
  }

  depends_on = [module.eks]
}

resource "kubernetes_config_map" "service_config" {
  metadata {
    name      = "service-config"
    namespace = kubernetes_namespace.finance.metadata[0].name
  }

  data = {
    SEED_GRPC_ADDR      = "seed:50051"
    ANALYTICS_GRPC_ADDR = "analytics:50052"
    ALERT_GRPC_ADDR     = "alert:50053"
    ENVIRONMENT         = var.environment
    AWS_REGION          = var.aws_region
  }
}
