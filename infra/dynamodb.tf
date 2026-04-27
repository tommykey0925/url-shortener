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

resource "aws_dynamodb_table" "urls_stats" {
  name         = "${var.project}-stats"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "code"
  range_key    = "date"

  attribute {
    name = "code"
    type = "S"
  }

  attribute {
    name = "date"
    type = "S"
  }

  tags = {
    Project = var.project
  }
}
