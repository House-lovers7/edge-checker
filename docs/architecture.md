# アーキテクチャ

edge-checker の内部構成と設計思想を解説します。

---

## 設計原則

1. **判定ツールであること** — 単なるリクエスト送信ツールではなく「防御が効いたか」を判定する検証器
2. **暴走しない設計** — 安全装置を最初から組み込み、意図しない大量リクエストを防止
3. **CLI-first** — GUI は後回し。CLI の方が速く、壊れにくく、Git 管理や CI にも乗る
4. **最小依存** — 外部依存は 4 つのみ。標準ライブラリを最大限活用

---

## レイヤ構成

```
┌──────────────────────────────────────────────┐
│  CLI (cobra)                                  │
│  cmd/edge-checker/main.go                     │
│  internal/cli/                                │
│    validate / dry-run / run / show / report    │
├──────────────────────────────────────────────┤
│  Scenario Layer                               │
│  internal/scenario/                           │
│    model.go   — YAML 構造体定義               │
│    loader.go  — ファイル読み込み + デフォルト適用 │
│    validator.go — バリデーションロジック         │
├──────────────────────────────────────────────┤
│  Safety Layer                                 │
│  internal/safety/                             │
│    allowlist.go — ホスト許可リスト              │
│    limits.go    — ハードリミットチェック         │
│    confirm.go   — production 環境確認          │
│    dns.go       — DNS 事前解決                 │
├──────────────────────────────────────────────┤
│  Execution Layer                              │
│  internal/engine/     internal/httpclient/     │
│    engine.go (IF)       client.go              │
│    constant.go          — HTTP 送信            │
│    burst.go             — レスポンス記録        │
│    ramp.go              — リダイレクト無効化     │
│    cooldown.go                                │
├──────────────────────────────────────────────┤
│  Observation Layer                            │
│  internal/observe/                            │
│    collector.go  — スレッドセーフなメトリクス収集 │
│    timeline.go   — 秒単位バケット              │
│    progress.go   — ターミナル進捗表示           │
├──────────────────────────────────────────────┤
│  Judgment Layer                               │
│  internal/judge/                              │
│    judge.go  — 統合判定 + Verdict 構造体        │
│    (block detection, block ratio)              │
├──────────────────────────────────────────────┤
│  Output Layer                                 │
│  internal/output/                             │
│    result.go    — Result 構造体               │
│    json.go      — JSON 保存                   │
│    markdown.go  — Markdown レポート生成        │
├──────────────────────────────────────────────┤
│  Profile Layer                                │
│  internal/profile/                            │
│    profile.go   — Profile 構造体              │
│    builtin.go   — 3 組み込みプロファイル        │
└──────────────────────────────────────────────┘
```

---

## データフロー

### run コマンドの実行フロー

```
YAML ファイル
    │
    ▼
┌─────────┐
│ Loader  │  → Scenario 構造体に変換
└────┬────┘
     ▼
┌──────────┐
│Validator │  → 構文・意味検証
└────┬─────┘
     ▼
┌─────────┐
│ Safety  │  → ホスト許可、上限、production、DNS チェック
└────┬────┘
     ▼
┌──────────┐
│ Profile  │  → ヘッダセット解決
└────┬─────┘
     ▼
┌──────────┐     ┌────────────┐
│  Engine  │────▶│ HTTPClient │────▶ ターゲット
│(constant │     └──────┬─────┘
│ burst    │            │ Response
│ ramp     │            ▼
│ cooldown)│     ┌────────────┐
└──────────┘     │ Collector  │  ← スレッドセーフに記録
                 │ + Timeline │
                 └──────┬─────┘
                        ▼
                 ┌────────────┐
                 │   Judge    │  → pass/fail 判定
                 └──────┬─────┘
                        ▼
                 ┌────────────┐
                 │  Output    │  → JSON + Markdown 保存
                 └────────────┘
```

---

## 並行制御の設計

### RPS 制御 + 並列数制御

Engine 内では 2 つのメカニズムが独立して動作します。

```go
limiter := rate.NewLimiter(rate.Limit(rps), rps)  // RPS 制御
sem := make(chan struct{}, concurrency)             // 並列数制御

for {
    limiter.Wait(ctx)     // 1. レートリミッターで待機
    sem <- struct{}{}     // 2. セマフォを取得
    go func() {
        defer func() { <-sem }()  // 3. セマフォを解放
        resp := client.Do(ctx, method, url)
        collector.Record(resp)     // 4. スレッドセーフに記録
    }()
}
```

- **rate.Limiter**: トークンバケットアルゴリズムで RPS を制御。`SetLimit()` で動的変更可能（burst/ramp で使用）
- **Channel セマフォ**: バッファ付きチャネルで同時 goroutine 数を制限
- **context**: 全体のタイムアウト・キャンセルを一元管理

### メトリクス収集のスレッドセーフ設計

```go
type Collector struct {
    mu        sync.Mutex       // 排他制御
    responses []Response       // 全レスポンス記録
    timeline  *Timeline        // 秒単位バケット
}
```

`sync.Mutex` で保護。1,000 rps 程度では mutex のコンテンション（競合）はボトルネックにならない（lock/unlock は数十ナノ秒）。

---

## 安全装置の設計

安全装置は **多層防御** で設計されています。

```
Layer 1: YAML Validation
  └─ 必須フィールド、型、制約の検証

Layer 2: Safety Checks (実行前)
  ├─ ホスト許可リスト
  ├─ ハードリミット（コード内定数、設定で緩和不可）
  ├─ 推定リクエスト数チェック
  ├─ production 環境の明示確認
  └─ DNS 事前解決

Layer 3: Runtime Guards (実行中)
  ├─ context.WithTimeout による全体タイムアウト
  ├─ MaxRequests によるリクエスト数上限
  └─ SIGINT/SIGTERM による graceful shutdown

Layer 4: Output Safety (実行後)
  └─ 中断時は interrupted: true を記録、判定は SKIP
```

### ハードリミット

```go
const (
    HardMaxTotalRequests = 100_000
    HardMaxConcurrency   = 100
    HardMaxDuration      = 30 * time.Minute
    HardMaxRPS           = 1_000
)
```

これらはコード内の定数であり、YAML や CLI フラグでは緩和できません。

---

## 判定ロジックの設計

Judge は **ルールベース** です。各ルールが独立して PASS/FAIL/SKIP を返し、1 つでも FAIL があれば全体が FAIL になります。

```go
type Verdict struct {
    Overall VerdictStatus   // PASS | FAIL | SKIP
    Rules   []RuleResult    // 各ルールの結果
}
```

### 現在のルール

| ルール | 判定内容 |
|--------|---------|
| Block detection | タイムラインをスキャンし、block_status_codes が block_within_seconds 以内に出現したか |
| Block ratio | blocked_count / total_count >= min_block_ratio か |

### 将来追加予定のルール

| ルール | 判定内容 |
|--------|---------|
| False positive | unaffected_paths がブロックされていないか |
| Bypass | bypass_scenarios がブロックされていないか |
| Cooldown release | Penalty Box が解除されたか |

---

## Engine interface の拡張

新しい実行モードを追加する場合:

1. `internal/engine/` に新しいファイルを作成
2. `Engine` interface を実装:
   ```go
   type Engine interface {
       Run(ctx context.Context, client *httpclient.Client, collector *observe.Collector) error
   }
   ```
3. `engine.New()` の switch に新しい case を追加
4. `scenario.model.go` に必要な設定構造体を追加
5. `scenario.validator.go` にバリデーションを追加

---

## 外部依存の選定理由

| ライブラリ | 選定理由 |
|-----------|---------|
| `cobra` | 6 サブコマンドの管理、PersistentFlags、自動ヘルプ生成。Go CLI のデファクト |
| `yaml.v3` | Go の YAML パース標準。構造体タグでマッピング |
| `x/time/rate` | トークンバケット。`SetLimit()` で動的変更可能（burst/ramp に必須） |
| `fatih/color` | PASS/FAIL の色付き表示。`--no-color` 対応。軽量 |

標準ライブラリで代替できないものだけを外部依存にしています。
