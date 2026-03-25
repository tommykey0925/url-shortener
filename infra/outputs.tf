output "cluster_name" {
  value = module.eks.cluster_name
}

output "cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "ecr_repository_url" {
  value = aws_ecr_repository.api.repository_url
}

output "dynamodb_table_name" {
  value = aws_dynamodb_table.urls.name
}

output "region" {
  value = var.region
}

output "api_role_arn" {
  value = module.irsa_dynamodb.iam_role_arn
}

output "cloudfront_distribution_id" {
  value = aws_cloudfront_distribution.frontend.id
}

output "cloudfront_domain_name" {
  value = aws_cloudfront_distribution.frontend.domain_name
}

output "frontend_bucket_name" {
  value = aws_s3_bucket.frontend.id
}
