# URL Shortener

長い URL を短縮し、クリック数を計測する Web アプリケーション。
AWS のインフラ構成・運用スキルを示すポートフォリオプロジェクト。

## Architecture

![Architecture](docs/architecture.png)

> [draw.io で開く](docs/architecture.drawio)

## Network

![Network](docs/network.png)

> [draw.io で開く](docs/network.drawio)

## CI/CD

![CI/CD](docs/cicd.png)

> [draw.io で開く](docs/cicd.drawio)

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | SvelteKit 2 + Svelte 5 + Tailwind CSS v4 |
| Backend | Go (net/http + AWS SDK v2) |
| Database | DynamoDB |
| Container | Docker → ECR → EKS (Kubernetes) |
| IaC | Terraform |
| GitOps | ArgoCD |
| CI/CD | GitHub Actions + flox |
| CDN | CloudFront (S3 + ALB dual origin) |

## Project Structure

```
url-shortener/
├── api/          # Go API
├── web/          # SvelteKit frontend
├── infra/        # Terraform
├── manifests/    # K8s manifests + ArgoCD
├── docs/         # Architecture diagrams (draw.io)
└── .github/      # CI/CD workflows
```

## API Endpoints

| Method | Path | Description |
|--------|------|------------|
| POST | `/api/shorten` | Shorten a URL |
| GET | `/api/urls` | List all URLs |
| GET | `/api/urls/{code}` | Get URL details |
| DELETE | `/api/urls/{code}` | Delete a URL |
| GET | `/r/{code}` | Redirect to original URL |
| GET | `/health` | Health check |

## Getting Started

### Prerequisites

```bash
# Install flox (manages all dev tools)
nix profile install --accept-flake-config github:flox/flox
```

### Local Development

```bash
flox activate                    # go, terraform, kubectl, pnpm, etc.
cd api && go run . &             # Start API on :8080
cd web && pnpm install && pnpm dev  # Start frontend on :5173
```

Vite dev server proxies `/api/*` and `/r/*` to the Go API.

### Deploy to AWS

```bash
cd infra && terraform apply      # Create EKS, DynamoDB, ECR, S3, CloudFront
# See CLAUDE.md for full deployment steps
```

### Tear Down

```bash
cd infra && terraform destroy    # Remove all AWS resources
```

## Cost

| Service | Monthly Cost |
|---------|-------------|
| EKS Control Plane | $73 |
| EC2 Nodes (t3.medium × 2) | ~$60 |
| ALB | ~$20 |
| Others (DynamoDB, S3, CloudFront) | ~$5 |
| **Total** | **~$158** |

> Cost strategy: `terraform destroy` when not in use, `terraform apply` for demos only.

## Security

- Rate limiting: 10 requests/min per IP
- DynamoDB: Provisioned capacity (read: 2/s, write: 1/s)
- S3: Private bucket, CloudFront OAC access only
- IAM: IRSA for pod-level permissions (least privilege)
