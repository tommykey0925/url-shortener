# Access Patterns

url-shortener は 2 テーブルで合計 **8 アクセスパターン** (IncrementClicks が両テーブルに 1 回ずつ更新するため)。

## 一覧

| # | Use case | Go method | Table | PK | SK / Filter | Source |
|---|---|---|---|---|---|---|
| 1 | URL を新規登録 | `Put` | `url-shortener` | `code` | - | [store.go:56](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L56) |
| 2 | code 単発取得 | `Get` | `url-shortener` | `code` (eq) | - | [store.go:81](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L81) |
| 3 | クリック数 +1 (URL 側) | `IncrementClicks` | `url-shortener` | `code` (eq) | `UpdateExpression: SET clicks = clicks + :inc` | [store.go:102](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L102) |
| 4 | 当日 daily stat に +1 | `IncrementClicks` | `url-shortener-stats` | `code` (eq) | `date = today` (eq); `SET clicks = if_not_exists(clicks, :zero) + :inc` | [store.go:118](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L118) |
| 5 | 過去 N 日の daily stats 取得 | `GetClickStats` | `url-shortener-stats` | `code` (eq) | `#d (= date) >= startDate` (Range); `ScanIndexForward=False` で降順 | [store.go:136](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L136) |
| 6 | safe_status 更新 (Lambda monitor) | `UpdateSafeStatus` | `url-shortener` | `code` (eq) | `UpdateExpression: SET safe_status = :s` | [store.go:162](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L162) |
| 7 | URL 削除 | `Delete` | `url-shortener` | `code` (eq) | `DeleteItem` | [store.go:176](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L176) |
| 8 | 全 URL 一覧 (admin) | `List` | `url-shortener` | (全件 Scan) | - | [store.go:186](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L186) |

## ScanIndexForward

`GetClickStats` (#5) は `ScanIndexForward=False` で SK 降順 = **新しい日付が先** ([store.go:149](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L149))。
これは `date` を SK に持つ設計が前提。

## DynamoDB 予約語の escape

`date` は DynamoDB の予約語のため、`GetClickStats` ([store.go:142](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L142)) では
`ExpressionAttributeNames: {"#d": "date"}` で escape している。`KeyConditionExpression` 内では `#d` 表記。

## Anti-patterns / Known concerns

### A1. `List` (#8) は全件 Scan
`store.go:186-194` で `Scan` をフィルタなしで実行。本番の `url-shortener` テーブル
は PROVISIONED `read=2` なので、URL 件数が増えると即枯渇するリスク。

- 改善案: 別 GSI を貼る (例: `created_at` index)、もしくは List API を admin 専用にしてクライアント側でページング
- 現状: 個人プロジェクトかつ件数小規模なので問題顕在化していない

### A2. `IncrementClicks` は **2 テーブル** の非トランザクション update
`store.go:103-115` (URL 側) と `store.go:118-129` (stats 側) は **2 つの独立した UpdateItem**。
片方だけ成功するケースが理論上あり得る (例: stats 側が ResourceNotFoundException など)。

- 改善案: `TransactWriteItems` で原子化 (2 テーブルに渡るので可能)、ただし PROVISIONED の場合 transactional は **2x コスト**
- 現状: 整合性ずれは「クリック数 +1 だが daily stats が遅れる」程度で実害低い

### A3. `IncrementClicks` の URL 側 update は初期値想定
`store.go:108` の `clicks = clicks + :inc` は `clicks` が `0` で初期化されている前提。
`Put` (`store.go:56`) 時に `Clicks: 0` を明示初期化しているので問題ないが、移行時は注意。

- 改善案: `clicks = if_not_exists(clicks, :zero) + :inc` の方が安全
- stats 側は既に `if_not_exists` を使っている ([store.go:124](https://github.com/tommykey-apps/url-shortener/blob/main/api/store/store.go#L124))

### A4. `UpdateSafeStatus` (#6) は ConditionExpression なし
`store.go:162-174` は `safe_status` を ブラインド上書き。Lambda の `monitor` から同時実行
されると順序保証なし。

- 改善案: optimistic locking (version 属性 + ConditionExpression)
- 現状: monitor は Lambda 単発で同時実行リスク低い

### A5. `Delete` (#7) は cascade なし
URL 本体を削除しても、対応する stats (`url-shortener-stats` の `code = X` の全 SK) は
削除されない。古い code を再利用した場合に古い stats が残ったまま IncrementClicks される。

- 改善案: Delete 内で stats の `Query` + `BatchDeleteItem` を呼ぶ、または TTL を利用
- 現状: code は十分にエントロピー高く再利用しない設計のため問題顕在化していない
