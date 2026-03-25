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

variable "alb_dns_name" {
  description = "DNS name of the ALB created by the ingress controller"
  type        = string
  default     = "placeholder.ap-northeast-1.elb.amazonaws.com"
}
