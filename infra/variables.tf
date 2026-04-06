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

variable "google_safe_browsing_api_key" {
  description = "Google Safe Browsing API key"
  type        = string
  sensitive   = true
  default     = ""
}

variable "groq_api_key" {
  description = "Groq API key for AI safety prediction"
  type        = string
  sensitive   = true
  default     = ""
}
