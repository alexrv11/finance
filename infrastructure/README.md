# Infrastructure

Terraform-managed infrastructure for local development (kind) and AWS (EKS).

## Structure

```
infrastructure/terraform/
  modules/
    vpc/        — AWS VPC, subnets, NAT gateway
    eks/        — EKS cluster, managed node group, OIDC/IRSA
    ecr/        — ECR repositories + lifecycle policies
  environments/
    local/      — kind cluster for local development
    aws/        — full AWS deployment (VPC + EKS + ECR)
```

---

## Prerequisites

| Tool        | Version   | Purpose                      |
|-------------|-----------|------------------------------|
| Terraform   | >= 1.6    | Infrastructure provisioning  |
| kind        | >= 0.23   | Local Kubernetes cluster     |
| Docker      | any       | Required by kind             |
| AWS CLI     | >= 2.x    | AWS auth + kubeconfig        |
| kubectl     | >= 1.30   | Cluster interaction          |

---

## Local Development

Spins up a 3-node kind cluster (1 control-plane + 2 workers) on your machine.

```bash
cd infrastructure/terraform/environments/local

terraform init
terraform apply

# Point kubectl at the local cluster
kind get kubeconfig --name finance-local > ~/.kube/config-finance-local
export KUBECONFIG=~/.kube/config-finance-local

# Verify
kubectl get nodes
kubectl get ns finance
```

NodePort mappings (configured in kind):
- `localhost:8080` → API Gateway (NodePort 30080)
- `localhost:9090` → Metrics / Prometheus (NodePort 30090)

To tear down:
```bash
terraform destroy
```

---

## AWS Deployment

Provisions a VPC, EKS cluster, managed node group, and ECR repositories.

### First-time setup

1. **Configure AWS credentials:**
   ```bash
   aws configure   # or export AWS_PROFILE=...
   ```

2. **Create remote state resources** (once per AWS account):
   ```bash
   aws s3api create-bucket \
     --bucket finance-terraform-state \
     --region us-east-1

   aws dynamodb create-table \
     --table-name finance-terraform-locks \
     --attribute-definitions AttributeName=LockID,AttributeType=S \
     --key-schema AttributeName=LockID,KeyType=HASH \
     --billing-mode PAY_PER_REQUEST \
     --region us-east-1
   ```

3. **Uncomment the S3 backend** in `environments/aws/backend.tf`.

### Deploy

```bash
cd infrastructure/terraform/environments/aws

cp terraform.tfvars.example terraform.tfvars
# edit terraform.tfvars as needed

terraform init
terraform plan
terraform apply
```

### Configure kubectl

```bash
aws eks update-kubeconfig --region us-east-1 --name finance-staging
kubectl get nodes
kubectl get ns finance
```

### Push an image to ECR

```bash
# Get the registry URL from Terraform output
ECR_URL=$(terraform output -json ecr_repository_urls | jq -r '.seed')

aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin "$ECR_URL"

docker build -t "$ECR_URL:latest" services/seed/
docker push "$ECR_URL:latest"
```

---

## Module Reference

### `modules/vpc`

| Input            | Type           | Default        | Description               |
|------------------|----------------|----------------|---------------------------|
| `name`           | string         | —              | Resource name prefix      |
| `cidr`           | string         | `10.0.0.0/16`  | VPC CIDR                  |
| `azs`            | list(string)   | —              | Availability zones        |
| `public_subnets` | list(string)   | —              | Public subnet CIDRs       |
| `private_subnets`| list(string)   | —              | Private subnet CIDRs      |

### `modules/eks`

| Input                | Type         | Default        | Description              |
|----------------------|--------------|----------------|--------------------------|
| `cluster_name`       | string       | —              | EKS cluster name         |
| `kubernetes_version` | string       | `1.30`         | K8s version              |
| `node_instance_types`| list(string) | `["t3.medium"]`| Worker instance types    |
| `node_desired_size`  | number       | `2`            | Desired node count       |

### `modules/ecr`

| Input      | Type         | Default  | Description                      |
|------------|--------------|----------|----------------------------------|
| `project`  | string       | —        | Repo prefix (`project/service`)  |
| `services` | list(string) | —        | One repo per service name        |
