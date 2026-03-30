# URL Monitor Lambda - Daily Safe Browsing re-check
resource "aws_lambda_function" "monitor" {
  function_name = "${var.project}-monitor"
  role          = aws_iam_role.monitor.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["x86_64"]
  timeout       = 60
  memory_size   = 128

  filename         = "${path.module}/lambda-monitor-placeholder.zip"
  source_code_hash = filebase64sha256("${path.module}/lambda-monitor-placeholder.zip")

  environment {
    variables = {
      DYNAMODB_TABLE               = aws_dynamodb_table.urls.name
      AWS_REGION_APP               = var.region
      GOOGLE_SAFE_BROWSING_API_KEY = var.google_safe_browsing_api_key
    }
  }

  tags = {
    Project = var.project
  }
}

# IAM Role for Monitor Lambda
resource "aws_iam_role" "monitor" {
  name = "${var.project}-monitor"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Project = var.project
  }
}

resource "aws_iam_role_policy_attachment" "monitor_basic" {
  role       = aws_iam_role.monitor.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "monitor_dynamodb" {
  name = "${var.project}-monitor-dynamodb"
  role = aws_iam_role.monitor.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:Scan",
          "dynamodb:UpdateItem",
        ]
        Resource = aws_dynamodb_table.urls.arn
      }
    ]
  })
}

# EventBridge Rule - Daily at 00:00 UTC (09:00 JST)
resource "aws_cloudwatch_event_rule" "daily_monitor" {
  name                = "${var.project}-daily-monitor"
  schedule_expression = "cron(0 0 * * ? *)"

  tags = {
    Project = var.project
  }
}

resource "aws_cloudwatch_event_target" "monitor" {
  rule = aws_cloudwatch_event_rule.daily_monitor.name
  arn  = aws_lambda_function.monitor.arn
}

resource "aws_lambda_permission" "eventbridge" {
  statement_id  = "AllowEventBridgeInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.monitor.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.daily_monitor.arn
}
