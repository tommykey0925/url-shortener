Destroy all AWS infrastructure for the URL shortener to save costs.

1. Check AWS credentials: `aws sts get-caller-identity`
2. Confirm with the user before proceeding (this will delete everything)
3. Run `cd infra && terraform destroy -auto-approve`
4. Verify all resources are cleaned up
5. Report estimated savings

Use `flox activate` before running commands. Warn the user that this will destroy EKS, DynamoDB data, S3 contents, and CloudFront distribution.
