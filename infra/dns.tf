data "aws_route53_zone" "main" {
  name = "tommykeyapp.com"
}

data "aws_acm_certificate" "wildcard" {
  provider = aws.us_east_1
  domain   = "*.tommykeyapp.com"
  statuses = ["ISSUED"]
}

# url.tommykeyapp.com -> CloudFront
resource "aws_route53_record" "url_shortener" {
  zone_id = data.aws_route53_zone.main.zone_id
  name    = "url.tommykeyapp.com"
  type    = "A"

  alias {
    name                   = aws_cloudfront_distribution.frontend.domain_name
    zone_id                = aws_cloudfront_distribution.frontend.hosted_zone_id
    evaluate_target_health = false
  }
}
