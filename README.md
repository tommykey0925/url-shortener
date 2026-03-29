# URL Shortener

URLを短くしてクリック数も見れるやつ。AWS使って色々やってみたくて作った。

## 構成図

![Architecture](docs/architecture.svg)

> [docs/architecture.drawio](docs/architecture.drawio) を draw.io で開くと編集できる

## 使った技術

| | |
|---|---|
| フロント | SvelteKit 2, Svelte 5, Tailwind CSS v4 |
| バックエンド | Go (標準ライブラリ + AWS SDK v2) |
| DB | DynamoDB |
| コンピュート | Lambda (API Gateway HTTP API) |
| IaC | Terraform |
| CI/CD | GitHub Actions + flox |
| 配信 | CloudFront (S3 + API Gateway を同一ドメインで配信) |

## 使ってるAWSサービス

| サービス | このプロジェクトでの役割 |
|---------|----------------------|
| Lambda | GoのAPIをサーバーレスで実行 |
| API Gateway | HTTPリクエストをLambdaにルーティング |
| DynamoDB | 短縮URLの保存先。NoSQLなのでキー検索が速い |
| S3 | フロントのビルド成果物を置いてる |
| CloudFront | CDN。S3とAPI Gatewayの前に立ってHTTPS配信 |
| Route 53 | カスタムドメイン (url.tommykeyapp.com) のDNS管理 |
| ACM | SSL証明書 (*.tommykeyapp.com ワイルドカード) |
| IAM | LambdaにDynamoDBアクセス権限を付与 |

## ディレクトリ構成

```
url-shortener/
├── api/          # GoのAPI（Lambda対応）
├── web/          # SvelteKitのフロント
├── infra/        # Terraform
├── docs/         # 構成図 (draw.io)
└── .github/      # GitHub Actionsのワークフロー
```

## APIの仕様

| Method | Path | 何するか |
|--------|------|---------|
| POST | `/api/shorten` | URLを短縮 |
| GET | `/api/urls` | 一覧取得 |
| GET | `/api/urls/{code}` | 1件取得 |
| DELETE | `/api/urls/{code}` | 削除 |
| GET | `/r/{code}` | 元のURLにリダイレクト |
| GET | `/health` | ヘルスチェック |

## ローカルで動かす

開発ツールは全部floxで管理してるので、まずfloxを入れる。

```bash
nix profile install --accept-flake-config github:flox/flox
```

あとはactivateすればgo, terraform, pnpm等が全部使える。

```bash
flox activate
cd api && go run . &                # APIが :8080 で起動
cd web && pnpm install && pnpm dev  # フロントが :5173 で起動
```

Viteのプロキシ設定で `/api/*` と `/r/*` がGoのAPIに流れるようにしてある。

## AWSにデプロイ

```bash
cd infra && terraform apply   # Lambda, API Gateway, DynamoDB, S3, CloudFront 等を作成
```

使い終わったら忘れずに壊す。

```bash
cd infra && terraform destroy
```

## コストについて

サーバーレス構成なので、アクセスがなければほぼ $0。
DynamoDBのプロビジョニングモード（読み2/秒・書き1/秒）で月数ドル程度。

## セキュリティ面

- Go側でIPあたり10リクエスト/分のレート制限をかけてる
- DynamoDBはプロビジョニングモードにして、読み2/秒・書き1/秒に抑えてる
- S3はプライベートバケットで、CloudFrontのOAC経由でしかアクセスできない
- LambdaのIAMロールは最小権限（DynamoDBのCRUD操作のみ）
