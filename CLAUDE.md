# URL Shortener

AWS ポートフォリオプロジェクト — Go API + SvelteKit + EKS + ArgoCD

## プロジェクト構成

```
url-shortener/
├── api/          # Go API (net/http + AWS SDK v2 + DynamoDB)
├── web/          # SvelteKit フロント (Svelte 5 + Tailwind v4 + adapter-static)
├── infra/        # Terraform (VPC, EKS, DynamoDB, ECR, S3, CloudFront, IRSA, ALB Controller)
├── manifests/    # K8s マニフェスト (Kustomize) + ArgoCD Application
├── docs/         # アーキテクチャ図 (draw.io)
└── .github/      # CI/CD (GitHub Actions + flox)
```

## 開発環境

**flox を使う。** 手動でツールをインストールしない。

```bash
flox activate   # go, terraform, kubectl, helm, argocd, eksctl, awscli, nodejs, pnpm, kustomize
```

## パッケージマネージャ

- Go: `go mod`
- Web (SvelteKit): **pnpm** (npm は使わない)

## コマンド

### API
```bash
cd api && go run .              # ローカル起動 (port 8080)
cd api && go test ./...         # テスト
cd api && go vet ./...          # 静的解析
cd api && docker build -t url-shortener-api .
```

### Web
```bash
cd web && pnpm install          # 依存インストール
cd web && pnpm dev              # ローカル起動 (Vite proxy で API 連携)
cd web && pnpm build            # ビルド (build/ に出力)
cd web && pnpm check            # 型チェック
```

### Infra
```bash
cd infra && terraform init
cd infra && terraform plan
cd infra && terraform apply     # デプロイ (EKS ~$150/月)
cd infra && terraform destroy   # 撤去 (使わないときは必ず destroy)
```

## API エンドポイント

| メソッド | パス | 機能 |
|---------|------|------|
| POST | `/api/shorten` | URL短縮 |
| GET | `/api/urls` | 一覧取得 |
| GET | `/api/urls/{code}` | 詳細取得 |
| DELETE | `/api/urls/{code}` | 削除 |
| GET | `/r/{code}` | リダイレクト (クリック数カウント) |
| GET | `/health` | ヘルスチェック |

## デプロイフロー

1. `terraform apply` → AWS インフラ構築
2. Docker build → ECR push
3. `kubectl` で ArgoCD インストール
4. ArgoCD が manifests/base/ を監視 → 自動デプロイ
5. `aws s3 sync web/build/ s3://bucket` → フロント配信

## 重要な設計判断

- リダイレクトパスは `/r/{code}` (CloudFront のルーティング衝突回避)
- CloudFront は S3 + ALB のデュアルオリジン構成 (CORS 不要)
- コスト節約: 使わないときは `terraform destroy`、デモ時だけ `terraform apply`
- GitOps: ArgoCD が manifests リポジトリの変更を検知して自動デプロイ

## DB スキーマドキュメント

`docs/db/` に DynamoDB スキーマドキュメント。`make db-docs` で再生成。
詳細: [docs/db/entities.md](docs/db/entities.md), [docs/db/access-patterns.md](docs/db/access-patterns.md)

## AWS リージョン

ap-northeast-1 (東京)
