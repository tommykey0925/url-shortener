variable "region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
}

variable "project" {
  description = "Project name"
  type        = string
  default     = "url-shortener"
}

variable "cluster_version" {
  description = "EKS cluster version"
  type        = string
  default     = "1.32"
}

data "aws_lb" "api" {
  tags = {
    "ingress.k8s.aws/stack" = "url-shortener/url-shortener-api"
  }
}
