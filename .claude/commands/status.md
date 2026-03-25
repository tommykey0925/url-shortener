Check the status of the URL shortener deployment on AWS.

1. Check AWS credentials: `aws sts get-caller-identity`
2. Check EKS cluster status: `aws eks describe-cluster --name url-shortener-cluster --query 'cluster.status'`
3. Check pods: `kubectl get pods -n url-shortener`
4. Check services: `kubectl get svc -n url-shortener`
5. Check ingress (ALB): `kubectl get ingress -n url-shortener`
6. Check ArgoCD app status: `argocd app get url-shortener` (or `kubectl get application -n argocd`)
7. Check CloudFront distribution status
8. Check DynamoDB table status
9. Report the CloudFront URL for accessing the app

Use `flox activate` before running commands. If any resource is not found, report it clearly.
