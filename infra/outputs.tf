output "cluster_name" {
  value = module.eks.cluster_name
}

output "cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "dynamodb_table_name" {
  value = aws_dynamodb_table.urls.name
}

output "region" {
  value = var.region
}

output "lambda_function_name" {
  value = aws_lambda_function.api.function_name
}

output "api_gateway_url" {
  value = aws_apigatewayv2_stage.default.invoke_url
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
