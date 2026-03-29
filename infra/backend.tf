terraform {
  backend "s3" {
    bucket         = "tommykeyapp-tfstate"
    key            = "url-shortener/terraform.tfstate"
    region         = "ap-northeast-1"
    dynamodb_table = "terraform-locks"
    encrypt        = true
  }
}
