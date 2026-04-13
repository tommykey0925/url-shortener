resource "aws_ssm_parameter" "cloudfront_distribution_id" {
  name  = "/${var.project}/cloudfront-distribution-id"
  type  = "String"
  value = aws_cloudfront_distribution.frontend.id
}

resource "aws_ssm_parameter" "frontend_bucket" {
  name  = "/${var.project}/frontend-bucket"
  type  = "String"
  value = aws_s3_bucket.frontend.bucket
}
