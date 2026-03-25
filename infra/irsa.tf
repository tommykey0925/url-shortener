data "aws_caller_identity" "current" {}

# IAM role for the API pods to access DynamoDB
module "irsa_dynamodb" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name = "${var.project}-api-dynamodb"

  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["${var.project}:${var.project}-api"]
    }
  }

  role_policy_arns = {
    dynamodb = aws_iam_policy.dynamodb_access.arn
  }
}

resource "aws_iam_policy" "dynamodb_access" {
  name = "${var.project}-dynamodb-access"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Scan",
        ]
        Resource = aws_dynamodb_table.urls.arn
      }
    ]
  })
}
