resource "aws_dynamodb_table" "urls" {
  name         = var.project
  billing_mode   = "PROVISIONED"
  hash_key       = "code"
  read_capacity  = 2
  write_capacity = 1

  attribute {
    name = "code"
    type = "S"
  }

  tags = {
    Project = var.project
  }
}
