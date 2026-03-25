resource "aws_dynamodb_table" "urls" {
  name         = var.project
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "code"

  attribute {
    name = "code"
    type = "S"
  }

  tags = {
    Project = var.project
  }
}
