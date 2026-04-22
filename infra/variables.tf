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
  description = "Google Safe Browsing API key (set via TF_VAR_google_safe_browsing_api_key from CD)"
  type        = string
  sensitive   = true
}

variable "groq_api_key" {
  description = "Groq API key for AI summarization (set via TF_VAR_groq_api_key from CD)"
  type        = string
  sensitive   = true
}
