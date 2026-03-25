Deploy the URL shortener to AWS. Run the following steps in order:

1. Check AWS credentials: `aws sts get-caller-identity`
2. Terraform apply (if infra not up): `cd infra && terraform apply -auto-approve`
3. Wait for EKS cluster to be ready
4. Configure kubectl: `aws eks update-kubeconfig --name url-shortener-cluster --region ap-northeast-1`
5. Build and push Docker image to ECR:
   - Get ECR URL from terraform output
   - `aws ecr get-login-password | docker login --username AWS --password-stdin <ecr-url>`
   - `cd api && docker build -t <ecr-url>:latest . && docker push <ecr-url>:latest`
6. Install ArgoCD on EKS (if not installed):
   - `kubectl create namespace argocd`
   - `kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml`
7. Update kustomization.yaml image tag
8. Apply ArgoCD application: `kubectl apply -f manifests/argocd/application.yaml`
9. Deploy frontend:
   - `cd web && pnpm build`
   - `aws s3 sync build/ s3://<bucket> --delete`
   - `aws cloudfront create-invalidation --distribution-id <id> --paths "/*"`
10. Show the CloudFront domain URL

Use `flox activate` before running any commands. Show progress at each step.
