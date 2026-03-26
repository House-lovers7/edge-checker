# edge-checker

WAF/CDN 防御設定の検証ハーネス CLI ツール。

自社の WAF/CDN（Akamai 等）防御ルールが想定どおりに動作するかを、再現可能な形で検証し、pass/fail 判定とレポートを出力します。

> **これは攻撃ツールではありません。**
> 許可済みホストに対してのみ動作する防御検証ツールです。
> 目的は「強い攻撃を打つこと」ではなく「どの設定が、どの条件で、どんな挙動を返すかを再現可能にすること」です。

## 目次

- [インストール](#インストール)
- [30秒クイックスタート](#30秒クイックスタート)
- [コマンド一覧](#コマンド一覧)
- [実行モード](#実行モード)
- [シナリオ YAML の書き方](#シナリオ-yaml-の書き方)
- [プロファイル](#プロファイル)
- [判定ロジック](#判定ロジック)
- [安全設計](#安全設計)
- [出力形式](#出力形式)
- [実務での使い方](#実務での使い方)
- [ディレクトリ構成](#ディレクトリ構成)
- [トラブルシューティング](#トラブルシューティング)
- [詳細ドキュメント](#詳細ドキュメント)
- [開発](#開発)

---

## インストール

### ソースからビルド（推奨）

```bash
git clone https://github.com/House-lovers7/edge-checker.git
cd edge-checker
make build
# ./edge-checker が生成される
```

### go install

```bash
go install github.com/House-lovers7/edge-checker/cmd/edge-checker@latest
```

### 動作確認

```bash
./edge-checker --version
# edge-checker dev (commit: xxxxxx, built: 2026-03-26T...)
```

---

## 30秒クイックスタート

```bash
# 1. シナリオの妥当性チェック
./edge-checker validate -f scenarios/assets-threshold.yaml

# 2. 実行計画の確認（リクエスト送信なし）
./edge-checker dry-run -f scenarios/assets-threshold.yaml

# 3. シナリオ実行
./edge-checker run -f scenarios/assets-threshold.yaml

# 4. 結果の確認
./edge-checker show -r results/result-XXXXXXXX-XXXXXX-assets-threshold-check.json

# 5. Markdown レポート生成
./edge-checker report -r results/result-XXXXXXXX-XXXXXX-assets-threshold-check.json
```

**推奨フロー: validate → dry-run → run の順に実行してください。**

dry-run で「何を、どれくらい、どこに打つか」を確認してから run に進むのが安全です。

---

## コマンド一覧

### validate — シナリオ検証

```bash
./edge-checker validate -f scenarios/assets-threshold.yaml
```

YAML の構文エラー、必須フィールドの欠落、モード別設定の整合性、allow_hosts とターゲットの一致をチェックします。

### dry-run — 実行計画の表示

```bash
./edge-checker dry-run -f scenarios/assets-burst.yaml
```

リクエストを一切送信せず、以下を表示します:
- ターゲット、モード、レート、プロファイル
- 推定総リクエスト数
- 安全チェック結果（ホスト許可、上限、環境）
- 判定基準

### run — シナリオ実行

```bash
./edge-checker run -f scenarios/assets-threshold.yaml
./edge-checker run -f scenarios/assets-threshold.yaml --allow-production
./edge-checker run -f scenarios/assets-threshold.yaml --insecure
./edge-checker run -f scenarios/assets-threshold.yaml -o custom-output.json
```

| フラグ | 説明 |
|-------|------|
| `-f <file>` | シナリオ YAML ファイル（必須） |
| `--allow-production` | `environment: production` のシナリオを実行許可 |
| `--insecure` | TLS 証明書検証をスキップ（自己署名証明書環境向け） |
| `-o <path>` | JSON 出力パスの上書き |

### show — 結果のターミナル表示

```bash
./edge-checker show -r results/result-20260326-103000-assets-threshold-check.json
```

### report — Markdown レポート生成

```bash
./edge-checker report -r results/result-20260326-103000-assets-threshold-check.json
./edge-checker report -r results/result-20260326-103000-assets-threshold-check.json -o report.md
```

### グローバルフラグ

| フラグ | 説明 |
|-------|------|
| `-v, --verbose` | 秒ごとの詳細出力 |
| `--no-color` | 色付き出力を無効化（パイプ、CI 向け） |
| `--version` | バージョン表示 |

---

## 実行モード

### constant — 一定レート

しきい値の前後を確認するための基本モード。

```yaml
execution:
  mode: constant
  duration: 30s
  concurrency: 5
rate:
  rps: 60
```

**用途:** Rate Policy のしきい値が何 rps で発火するかの確認、bypass IP との差分比較

### burst — ベースレート + 周期的スパイク

平常時のアクセスに加え、一定間隔で急増させるモード。

```yaml
execution:
  mode: burst
  duration: 60s
  concurrency: 10
rate:
  rps: 15              # ベースレート
  burst:
    spike_rps: 100      # スパイク時のレート
    spike_duration: 5s   # スパイクの持続時間
    interval: 15s        # スパイクの間隔
```

**用途:** バースト耐性の確認、Penalty Box 発火タイミングの特定

### ramp — 段階的増加

レートを徐々に増やして、しきい値の境界を探索するモード。

```yaml
execution:
  mode: ramp
  duration: 60s
  concurrency: 10
rate:
  rps: 10              # 初期レート（フォールバック値）
  ramp:
    start_rps: 5
    end_rps: 100
    step_duration: 10s   # 段階の間隔
```

**用途:** 「何 rps から引っかかるか」の境界探索

### cooldown-check — Penalty Box 解除確認

3フェーズ実行で、防御発火後の解除タイミングを確認するモード。

```
Phase 1 (Active)   → 高レートで打って防御を発火させる
Phase 2 (Wait)     → 低頻度プローブでブロック状態を観測する
Phase 3 (Verify)   → 中レートで復帰を確認する
```

```yaml
execution:
  mode: cooldown-check
  duration: 120s
  concurrency: 5
rate:
  rps: 60
  cooldown:
    active_duration: 15s
    active_rps: 80
    wait_duration: 60s
    probe_rps: 1
```

**用途:** Penalty Box が何分で解除されるか、解除後に正常応答が戻るかの確認

---

## シナリオ YAML の書き方

最小限の constant モードシナリオ:

```yaml
name: my-first-test
description: Rate limit threshold check

target:
  base_url: https://staging.example.com
  path: /assets/images/medium/sample.jpg

execution:
  mode: constant
  duration: 30s

rate:
  rps: 60

safety:
  allow_hosts:
    - staging.example.com

expect:
  block_status_codes: [403, 429]
```

全フィールドの詳細は [docs/scenario-reference.md](docs/scenario-reference.md) を参照してください。

### デフォルト値

| フィールド | デフォルト |
|-----------|----------|
| `target.method` | `GET` |
| `execution.timeout` | `10s` |
| `execution.concurrency` | `1` |
| `safety.environment` | `staging` |

### 出力パスのプレースホルダ

`output.json` と `output.markdown` で使えるプレースホルダ:

| プレースホルダ | 展開例 |
|--------------|--------|
| `{{timestamp}}` | `20260326-103000` |
| `{{name}}` | `assets-threshold-check` |

---

## プロファイル

ヘッダセットで「どういうクライアントに見えるか」を切り替えます。

| 名前 | User-Agent | Accept | 用途 |
|------|-----------|--------|------|
| `browser-like` | Chrome 125 | HTML+画像 | 正常ブラウザ相当 |
| `bot-like` | TestBot/1.0 | `*/*` | 最小限ボット |
| `crawler-like` | TestCrawler/1.0 | HTML | クローラ相当 |

```yaml
profile: bot-like
```

**設計意図:** 「本物の偽装」ではなく、WAF の判定に差が出る最小限の属性差を作ることが目的です。どの識別軸で反応が変わるかを切り分けるために使います。

シナリオの `headers` フィールドで追加ヘッダやプロファイルの上書きも可能です:

```yaml
profile: browser-like
headers:
  X-Custom-Header: test-value
  User-Agent: "Mozilla/5.0 (custom override)"  # プロファイルを上書き
```

---

## 判定ロジック

run 完了後、`expect` セクションに基づいて自動判定します。

### Block detection

指定秒数以内にブロックステータス（403, 429 等）が出現したか。

```yaml
expect:
  block_status_codes: [403, 429]
  block_within_seconds: 60
```

- PASS: 403 or 429 が 60 秒以内に出現
- FAIL: 出現しない、または遅すぎる

### Block ratio

ブロックされたリクエストの比率が閾値以上か。

```yaml
expect:
  min_block_ratio: 0.20  # 20% 以上
```

- PASS: ブロック率 >= 20%
- FAIL: ブロック率 < 20%

### 判定結果の見方

```
--- Verdict: PASS ---
  ✓ Block detection    : First block at 3s (expected: Block status [403 429] within 60s)
  ✓ Block ratio        : 32.5% (600/1845) (expected: >= 20%)
```

| アイコン | ステータス | 意味 |
|---------|----------|------|
| ✓ | PASS | 期待通り |
| ✗ | FAIL | 期待と不一致 |
| - | SKIP | 判定基準なし、または中断 |

---

## 安全設計

このツールは **暴走しない設計** を最優先にしています。

### 1. ホスト許可リスト

`safety.allow_hosts` にないホストへの実行は拒否されます。

```yaml
safety:
  allow_hosts:
    - staging.example.com
    - test.example.com
```

### 2. 本番保護

`environment: production` のシナリオは `--allow-production` フラグがないと実行できません。

```bash
# これはブロックされる
./edge-checker run -f scenarios/prod-check.yaml

# これは実行できる
./edge-checker run -f scenarios/prod-check.yaml --allow-production
```

### 3. ハードリミット

コード内の定数で以下を制限しています。シナリオ設定では緩和できません。

| 項目 | 上限 |
|------|------|
| 最大リクエスト数 | 100,000 |
| 最大並列数 | 100 |
| 最大 RPS | 1,000 |
| 最大実行時間 | 30 分 |

### 4. DNS 事前チェック

実行前にターゲットホストの DNS 解決を確認します。解決できない場合は警告を表示します。

### 5. Graceful Shutdown

`Ctrl+C` (SIGINT) または `SIGTERM` で安全に停止します。
- 実行中のリクエストの完了を待機
- 途中結果を JSON に保存（`"interrupted": true`）
- 判定は SKIP（中断のため信頼性なし）

### 6. dry-run

`dry-run` コマンドで実行計画を事前確認できます。本番系のシナリオでは必ず先に dry-run を実行してください。

---

## 出力形式

### JSON（機械処理用）

```json
{
  "scenario_name": "assets-threshold-check",
  "started_at": "2026-03-26T10:30:00+09:00",
  "ended_at": "2026-03-26T10:30:30+09:00",
  "duration": "30s",
  "summary": {
    "total_requests": 1800,
    "success_count": 1400,
    "status_counts": { "200": 1400, "403": 320, "429": 80 },
    "p95_latency_ms": 42.5
  },
  "timeline": [ ... ],
  "verdict": {
    "overall": "PASS",
    "rules": [ ... ]
  }
}
```

### Markdown（報告資料用）

`run` 時にシナリオの `output.markdown` を指定するか、`report` コマンドで後から生成します。

```bash
./edge-checker report -r results/result.json -o report.md
```

生成されるレポートにはシナリオ情報、サマリ、ステータス分布、Verdict テーブル、秒単位タイムラインが含まれます。そのまま社内報告や引き継ぎ資料に貼れます。

---

## 実務での使い方

### 1. 設定変更前の事前確認

Rate Policy を変更する前に、現在の設定が効いているかを確認。

```bash
./edge-checker run -f scenarios/assets-threshold.yaml
# → PASS なら現在の設定は有効
```

### 2. 設定変更後の回帰確認

Rate Policy を変更した後に、新しい設定が期待通りに動くかを確認。

```bash
# rps を変えた 2 パターンで検証
./edge-checker run -f scenarios/assets-threshold-low.yaml
./edge-checker run -f scenarios/assets-threshold-high.yaml
```

### 3. bypass IP の検証

同じシナリオを bypass IP と非 bypass IP から実行して差分を比較。

```bash
# 非 bypass IP から実行（ブロックされるべき）
./edge-checker run -f scenarios/bypass-compare.yaml -o results/non-bypass.json

# bypass IP から実行（ブロックされないべき）
./edge-checker run -f scenarios/bypass-compare.yaml -o results/bypass.json

# 両方の結果を比較
./edge-checker show -r results/non-bypass.json
./edge-checker show -r results/bypass.json
```

### 4. Penalty Box 解除時間の確認

```bash
./edge-checker run -f scenarios/assets-cooldown.yaml
# → Phase 1 で発火、Phase 2 でブロック状態を観測、Phase 3 で復帰確認
```

### 5. 報告資料の作成

```bash
./edge-checker report -r results/result.json -o report.md
# → そのまま社内チケットや Confluence に貼る
```

詳細は [docs/usage-guide.md](docs/usage-guide.md) を参照してください。

---

## ディレクトリ構成

```
edge-checker/
├── cmd/edge-checker/        # エントリーポイント
│   └── main.go
├── internal/
│   ├── cli/                 # Cobra コマンド定義
│   ├── engine/              # 実行モード (constant/burst/ramp/cooldown)
│   ├── httpclient/          # HTTP クライアント
│   ├── judge/               # pass/fail 判定ロジック
│   ├── observe/             # メトリクス収集・タイムライン・進捗表示
│   ├── output/              # JSON/Markdown 出力
│   ├── profile/             # ヘッダープロファイル
│   ├── safety/              # 安全装置 (allowlist/limits/DNS/production確認)
│   ├── scenario/            # YAML モデル・ローダー・バリデーション
│   └── version/             # バージョン情報
├── scenarios/               # サンプルシナリオ YAML
├── results/                 # 実行結果出力先
├── docs/                    # 詳細ドキュメント
├── Makefile
└── README.md
```

---

## トラブルシューティング

### `host "xxx" is not in allow_hosts`

ターゲット URL のホスト名が `safety.allow_hosts` に含まれていません。
allow_hosts にホスト名を追加してください（ポート番号含む場合はポートも一致させる）。

### `environment is "production" but --allow-production flag was not set`

本番環境へのシナリオです。意図的であれば `--allow-production` フラグを付けてください。

### `estimated total requests X exceeds max_total_requests Y`

推定リクエスト数がシナリオの `max_total_requests` を超えています。
rps を下げるか、duration を短くするか、max_total_requests を引き上げてください。

### `DNS resolution failed for "xxx"`

ホスト名が解決できません。ネットワーク接続やホスト名のスペルを確認してください。
DNS チェックは警告のみで実行は続行しますが、リクエストはすべてエラーになります。

### 全リクエストがエラーで status_counts が空

ターゲットホストに到達できていません。以下を確認:
- ホスト名が正しいか（DNS 解決できるか）
- VPN / プロキシが必要か
- ファイアウォールでブロックされていないか
- `--insecure` が必要か（自己署名証明書の場合）

### Ctrl+C しても止まらない

1 回目の Ctrl+C で graceful shutdown を開始します（実行中リクエストの完了を待機）。
2 回目の Ctrl+C で強制終了します。

---

## 詳細ドキュメント

| ドキュメント | 内容 |
|-------------|------|
| [docs/scenario-reference.md](docs/scenario-reference.md) | シナリオ YAML 全フィールドリファレンス |
| [docs/usage-guide.md](docs/usage-guide.md) | 実務ワークフロー詳細ガイド |
| [docs/architecture.md](docs/architecture.md) | 内部構成と設計思想 |

---

## 開発

```bash
make build    # バイナリビルド（ldflags でバージョン埋め込み）
make test     # テスト実行（race detector 有効）
make clean    # バイナリ削除
```

### 外部依存（4 つのみ）

| ライブラリ | 用途 |
|-----------|------|
| `github.com/spf13/cobra` | CLI サブコマンド |
| `gopkg.in/yaml.v3` | YAML パース |
| `golang.org/x/time/rate` | RPS 制御（トークンバケット） |
| `github.com/fatih/color` | ターミナル色付き出力 |

## ライセンス

Private
