# Entities

url-shortener は **2 テーブル** + **派生型 1**。短縮 URL のメタデータと日次クリック統計を分離して保持する。

## 一覧

| Entity | Table | PK / SK | Go struct | 主な関数 |
|---|---|---|---|---|
| URL | `url-shortener` | PK=`code` | [`URL`](https://github.com/tommykey-apps/url-shortener/blob/main/api/model/url.go#L5) | [`Put`](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L56), [`Get`](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L81), [`Delete`](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L176), [`List`](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L186) |
| DailyClicks | `url-shortener-stats` | PK=`code`, SK=`date` | [`DailyClicks`](https://github.com/tommykey-apps/url-shortener/blob/main/api/model/url.go#L27) | [`IncrementClicks`](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L102), [`GetClickStats`](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L136) |
| ClickStats (派生) | (DB に存在せず、API レスポンス用集計) | - | [`ClickStats`](https://github.com/tommykey-apps/url-shortener/blob/main/api/model/url.go#L32) | API handler が GetClickStats の結果を集計 |

---

## URL

短縮 URL の本体。

- **Table**: `url-shortener`
- **PK**: `code` (String)
- **SK**: なし
- **Billing**: 本番 PROVISIONED (read=2 / write=1)、ローカル / CI PAY_PER_REQUEST

| Field | Go type | DynamoDB type | dynamodbav tag | Source |
|---|---|---|---|---|
| `Code` | string | S | `code` | [model/url.go:6](https://github.com/tommykey-apps/url-shortener/blob/main/api/model/url.go#L6) |
| `Original` | string | S | `original_url` | [model/url.go:7](https://github.com/tommykey-apps/url-shortener/blob/main/api/model/url.go#L7) |
| `CreatedAt` | time.Time | S (ISO 8601) | `created_at` | [model/url.go:8](https://github.com/tommykey-apps/url-shortener/blob/main/api/model/url.go#L8) |
| `Clicks` | int64 | N | `clicks` | [model/url.go:9](https://github.com/tommykey-apps/url-shortener/blob/main/api/model/url.go#L9) |
| `SafeStatus` | string | S | `safe_status` | [model/url.go:10](https://github.com/tommykey-apps/url-shortener/blob/main/api/model/url.go#L10) |

サンプルアイテム:
```json
{
  "code": "abc123",
  "original_url": "https://example.com/very/long/path",
  "created_at": "2026-05-03T01:23:45Z",
  "clicks": 42,
  "safe_status": "SAFE"
}
```

---

## DailyClicks

日付ごとのクリック数の集計。`code` と `date` の複合キーで時系列クエリが可能。

- **Table**: `url-shortener-stats`
- **PK**: `code` (URL.code を参照、外部キー扱いだが DynamoDB は FK 制約なし)
- **SK**: `date` (YYYY-MM-DD)
- **Billing**: PAY_PER_REQUEST

| Field | Go type | DynamoDB type | dynamodbav tag | Source |
|---|---|---|---|---|
| `Date` | string | S (YYYY-MM-DD) | `date` | [model/url.go:28](https://github.com/tommykey-apps/url-shortener/blob/main/api/model/url.go#L28) |
| `Clicks` | int64 | N | `clicks` | [model/url.go:29](https://github.com/tommykey-apps/url-shortener/blob/main/api/model/url.go#L29) |

サンプルアイテム:
```json
{
  "code": "abc123",
  "date": "2026-05-03",
  "clicks": 7
}
```

**注意**: `date` は DynamoDB の予約語のため、`store.go:142` の Query では
`ExpressionAttributeNames` で `#d = "date"` として escape している。

---

## ClickStats (派生型)

DB に存在せず、`GetClickStats` の結果を集計して API レスポンスに使う型。

- Go struct: [`ClickStats`](https://github.com/tommykey-apps/url-shortener/blob/main/api/model/url.go#L32)

| Field | Go type | 由来 |
|---|---|---|
| `Code` | string | URL.code |
| `TotalClicks` | int64 | URL.clicks (累積総和) |
| `Daily` | []DailyClicks | GetClickStats の結果 |

API handler が `GetClickStats(code, days)` の結果と `Get(code).Clicks` を組み合わせて返す。

---

## 設計意図

- **2 テーブル分離**: URL 本体は read-heavy、stats は write-heavy (クリック毎に書き込み) なので別テーブルに分離してビリングモードも分ける
- **Stats の SK = date**: `begins_with(EXP#YYYY-MM)` 等での月別 / 期間別クエリを想定 (現状実装は `>= start` だが将来拡張可能)
- **ScanIndexForward = false**: GetClickStats は降順 (新しい日付が先) で返す ([store.go:149](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L149))
- **`if_not_exists` で初期 0 化**: stats の IncrementClicks は `clicks = if_not_exists(clicks, :zero) + :inc` ([store.go:124](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L124)) で「初日のクリック」も正しく 1 にカウント
