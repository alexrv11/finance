output "cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "EKS cluster API endpoint"
  value       = module.eks.cluster_endpoint
}

output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "ecr_repository_urls" {
  description = "Map of service → ECR repository URL"
  value       = module.ecr.repository_urls
}

output "oidc_provider_arn" {
  description = "OIDC provider ARN — used when creating IRSA IAM policies"
  value       = module.eks.oidc_provider_arn
}

output "kubeconfig_command" {
  description = "Run this to configure kubectl for this cluster"
  value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${module.eks.cluster_name}"
}
