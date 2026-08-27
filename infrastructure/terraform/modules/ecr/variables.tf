variable "project" {
  description = "Project name used as ECR repo prefix"
  type        = string
}

variable "services" {
  description = "List of microservice names — one ECR repo is created per service"
  type        = list(string)
}

variable "image_tag_mutability" {
  description = "Tag mutability: MUTABLE or IMMUTABLE"
  type        = string
  default     = "MUTABLE"
}

variable "scan_on_push" {
  description = "Enable image vulnerability scanning on push"
  type        = bool
  default     = true
}

variable "keep_image_count" {
  description = "Number of images to retain per repository"
  type        = number
  default     = 10
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
