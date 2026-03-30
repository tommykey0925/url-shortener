# URL Shortener

URLを短くしてクリック数も計測できるサービス。登録時にURLの安全性を自動チェックし、AIによる遷移先の要約機能も搭載。

## 構成図

![Architecture](docs/architecture.svg)

> [docs/architecture.drawio](docs/architecture.drawio) を draw.io で開くと編集できる

## 使った技術

| | |
|---|---|
| フロント | SvelteKit 2, Svelte 5, Tailwind CSS v4, shadcn-svelte |
| バックエンド | Go (標準ライブラリ + AWS SDK v2) |
| DB | DynamoDB |
| コンピュート | Lambda (API Gateway HTTP API) |
| AI | Groq API (Llama 3.3 70B) — URL遷移先の要約・安全性判定 |
| セキュリティ | Google Safe Browsing API, DNS解決チェック |
| IaC | Terraform (S3バックエンド + DynamoDB state lock) |
| CI/CD | GitHub Actions + flox (Terraform apply + Lambda deploy 自動化) |
| 配信 | CloudFront (S3 + API Gateway を同一ドメインで配信) |

## 使ってるAWSサービス

| サービス | このプロジェクトでの役割 |
|---------|----------------------|
| Lambda | GoのAPIをサーバーレスで実行 |
| API Gateway | HTTPリクエストをLambdaにルーティング |
| DynamoDB | 短縮URLの保存先。NoSQLなのでキー検索が速い |
| S3 | フロントのビルド成果物 + Terraform state の保存 |
| CloudFront | CDN。S3とAPI Gatewayの前に立ってHTTPS配信 |
| Route 53 | カスタムドメイン (url.tommykeyapp.com) のDNS管理 |
| ACM | SSL証明書 (*.tommykeyapp.com ワイルドカード) |
| IAM | LambdaにDynamoDBアクセス権限を付与 |

## セキュリティ・安全性チェック

URL登録時に3段階のチェックを実施:

1. **DNS解決チェック** — ドメインが実在するか検証。存在しないドメインは拒否
2. **Google Safe Browsing API** — Googleのブラックリストに照合。マルウェア・フィッシングサイトを拒否
3. **AI要約 (Groq / Llama 3.3 70B)** — 登録後に「AI要約」ボタンで遷移先を分析。ページのHTML（title, description, 本文冒頭）を取得してLLMに要約させる。怪しいサイトの場合は警告を表示

URL一覧はブラウザのlocalStorageで自分が作成したURLだけ表示される。

## ディレクトリ構成

```
url-shortener/
├── api/          # GoのAPI（Lambda対応）
│   └── safety/   # URL安全性チェック（Safe Browsing, AI, DNS）
├── web/          # SvelteKitのフロント（shadcn-svelte）
├── infra/        # Terraform（S3 backend）
├── docs/         # 構成図 (draw.io)
└── .github/      # GitHub Actions（Lambda deploy + Terraform apply）
```

## APIの仕様

| Method | Path | 何するか |
|--------|------|---------|
| POST | `/api/shorten` | URLを短縮（DNS + Safe Browsing チェック付き） |
| GET | `/api/urls` | 一覧取得 |
| GET | `/api/urls/{code}` | 1件取得 |
| DELETE | `/api/urls/{code}` | 削除 |
| POST | `/api/urls/{code}/summarize` | AIによるURL遷移先の要約 |
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

- Go側でIPあたり30リクエスト/分のレート制限をかけてる
- DynamoDBはプロビジョニングモードにして、読み2/秒・書き1/秒に抑えてる
- S3はプライベートバケットで、CloudFrontのOAC経由でしかアクセスできない
- LambdaのIAMロールは最小権限（DynamoDBのCRUD操作のみ）
- URL登録時にDNS解決 + Google Safe Browsing でチェック
