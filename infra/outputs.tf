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
