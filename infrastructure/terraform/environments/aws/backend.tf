# Remote state — uncomment and fill in before first apply.
# Create the S3 bucket and DynamoDB table manually once, then enable this block.
#
# terraform {
#   backend "s3" {
#     bucket         = "finance-terraform-state"          # must exist
#     key            = "aws/${var.environment}/terraform.tfstate"
#     region         = "us-east-1"
#     encrypt        = true
#     dynamodb_table = "finance-terraform-locks"          # for state locking
#   }
# }
