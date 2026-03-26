# シナリオ YAML リファレンス

edge-checker のシナリオ YAML の全フィールドを解説します。

---

## 全体構造

```yaml
name: string            # 必須
description: string     # 任意

target:                 # 必須
  base_url: string      # 必須 — https://staging.example.com
  host: string          # 任意 — Host ヘッダ上書き（CDN ルーティング用）
  path: string          # 必須 — /assets/images/medium/sample.jpg
  method: string        # 任意 — GET (デフォルト)

execution:              # 必須
  mode: string          # 必須 — constant | burst | ramp | cooldown-check
  duration: string      # 必須 — Go duration 形式 (30s, 1m30s, 5m)
  timeout: string       # 任意 — リクエストタイムアウト (デフォルト: 10s)
  concurrency: int      # 任意 — 同時接続数 (デフォルト: 1)

rate:                   # 必須
  rps: int              # 必須 — 基本リクエスト/秒
  burst: BurstConfig    # burst モード時に必須
  ramp: RampConfig      # ramp モード時に必須
  cooldown: CooldownConfig  # cooldown-check モード時に必須

profile: string         # 任意 — browser-like | bot-like | crawler-like

headers:                # 任意 — 追加ヘッダ (map[string]string)
  Key: Value

query:                  # 任意
  static: map           # 固定クエリパラメータ
  random_suffix: bool   # キャッシュバスティング用ランダムサフィックス

safety:                 # 必須
  environment: string   # 任意 — staging (デフォルト) | production
  max_total_requests: int  # 任意 — 最大リクエスト数上限
  allow_hosts: [string] # 必須 — 許可ホストリスト

expect:                 # 必須
  block_status_codes: [int]     # 必須 — ブロック判定対象ステータスコード
  block_within_seconds: int     # 任意 — 何秒以内にブロック出現を期待するか
  min_block_ratio: float        # 任意 — 最低ブロック率 (0.0〜1.0)
  unaffected_paths: [UnaffectedPath]   # 任意 — 誤検知チェック対象
  bypass_scenarios: [BypassScenario]   # 任意 — バイパスチェック対象

output:                 # 任意
  json: string          # JSON 出力パス (プレースホルダ使用可)
  markdown: string      # Markdown 出力パス (プレースホルダ使用可)
```

---

## フィールド詳細

### name

| | |
|---|---|
| 型 | string |
| 必須 | はい |
| 用途 | シナリオの識別名。結果ファイル名にも使用される |

```yaml
name: assets-threshold-check
```

### description

| | |
|---|---|
| 型 | string |
| 必須 | いいえ |
| 用途 | シナリオの目的説明。レポートに出力される |

---

## target

### target.base_url

| | |
|---|---|
| 型 | string |
| 必須 | はい |
| 制約 | `http://` または `https://` で始まること |
| 例 | `https://staging.example.com` |

ターゲットの URL（パスを除いたベース部分）。

### target.host

| | |
|---|---|
| 型 | string |
| 必須 | いいえ |
| 用途 | HTTP リクエストの `Host` ヘッダを上書き |

CDN 経由でテストする場合、CDN のホスト名を `base_url` に、本来のドメインを `host` に設定します。

```yaml
target:
  base_url: https://cdn-staging.example.com
  host: www.example.com    # CDN に「このドメイン宛て」と伝える
  path: /assets/sample.jpg
```

### target.path

| | |
|---|---|
| 型 | string |
| 必須 | はい |
| 例 | `/assets/images/medium/sample.jpg` |

`base_url` に連結されてリクエスト URL を構成します。

### target.method

| | |
|---|---|
| 型 | string |
| 必須 | いいえ |
| デフォルト | `GET` |
| 有効値 | `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `PATCH` |

---

## execution

### execution.mode

| | |
|---|---|
| 型 | string |
| 必須 | はい |
| 有効値 | `constant`, `burst`, `ramp`, `cooldown-check` |

実行モード。モードごとに `rate` セクションの必須フィールドが異なります。

| モード | rate で必須 |
|--------|-----------|
| `constant` | `rps` のみ |
| `burst` | `rps` + `burst.*` |
| `ramp` | `rps` + `ramp.*` |
| `cooldown-check` | `rps` + `cooldown.*` |

### execution.duration

| | |
|---|---|
| 型 | string (Go duration) |
| 必須 | はい |
| 例 | `30s`, `1m30s`, `5m` |

テスト全体の実行時間。cooldown-check の場合、各フェーズの合計ではなくタイムアウト上限として扱われます。

### execution.timeout

| | |
|---|---|
| 型 | string (Go duration) |
| 必須 | いいえ |
| デフォルト | `10s` |

個々の HTTP リクエストのタイムアウト。

### execution.concurrency

| | |
|---|---|
| 型 | int |
| 必須 | いいえ |
| デフォルト | `1` |
| ハードリミット | `100` |

同時に飛ばすリクエストの数。RPS 制御とは独立して、同時接続数を制限します。

---

## rate

### rate.rps

| | |
|---|---|
| 型 | int |
| 必須 | はい |
| ハードリミット | `1,000` |

1 秒あたりのリクエスト数。トークンバケットアルゴリズムで制御されます。

### rate.burst（burst モード用）

```yaml
rate:
  rps: 15                # ベースレート
  burst:
    spike_rps: 100       # スパイク時のレート
    spike_duration: 5s   # スパイクの持続時間
    interval: 15s        # スパイクの間隔
```

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `spike_rps` | int | スパイク時のリクエスト/秒 |
| `spike_duration` | string | スパイクの持続時間 |
| `interval` | string | スパイク発生の間隔 |

**動作:** interval ごとに spike_duration の間だけ spike_rps に切り替わり、その後 rps に戻る。

```
rps: |     100 ┤    ___         ___         ___
     |   15 ┤___/   \___/   \___/   \___
     |        0s  15s 20s  30s 35s  45s 50s
```

### rate.ramp（ramp モード用）

```yaml
rate:
  rps: 10
  ramp:
    start_rps: 5
    end_rps: 100
    step_duration: 10s
```

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `start_rps` | int | 開始時のリクエスト/秒 |
| `end_rps` | int | 最終的なリクエスト/秒 |
| `step_duration` | string | 段階の間隔 |

**動作:** step_duration ごとに均等に RPS を増加させる。

```
rps: |  100 ┤                        ____
     |   70 ┤                   ____/
     |   40 ┤              ____/
     |   10 ┤    _________/
     |    5 ┤___/
     |        0s  10s  20s  30s  40s  50s  60s
```

### rate.cooldown（cooldown-check モード用）

```yaml
rate:
  rps: 60
  cooldown:
    active_duration: 15s
    active_rps: 80
    wait_duration: 60s
    probe_rps: 1
```

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `active_duration` | string | Phase 1（攻撃フェーズ）の持続時間 |
| `active_rps` | int | Phase 1 のリクエスト/秒 |
| `wait_duration` | string | Phase 2（待機フェーズ）の持続時間 |
| `probe_rps` | int | Phase 2 のプローブリクエスト/秒 |

**動作:**
1. **Phase 1 (Active):** active_rps で active_duration だけ送信し、防御を発火させる
2. **Phase 2 (Wait):** probe_rps の低頻度プローブで wait_duration 待機し、ブロック状態を観測する
3. **Phase 3 (Verify):** active_rps の半分のレートで 5 秒間送信し、復帰を確認する

---

## profile

| | |
|---|---|
| 型 | string |
| 必須 | いいえ |
| 有効値 | `browser-like`, `bot-like`, `crawler-like` |

ヘッダプロファイル。未指定の場合はプロファイルなし（ヘッダ最小）。

---

## headers

| | |
|---|---|
| 型 | map[string]string |
| 必須 | いいえ |

追加ヘッダ。プロファイルで設定されたヘッダを上書きすることもできます。

```yaml
headers:
  X-Forwarded-For: "192.168.1.100"
  Referer: "https://www.example.com/"
```

---

## query

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `static` | map[string]string | 固定クエリパラメータ |
| `random_suffix` | bool | `_t=<random>` を付与（キャッシュバスティング） |

```yaml
query:
  static:
    format: webp
    quality: "80"
  random_suffix: true
```

> **注:** query 機能は現在 YAML 定義のみで、実行時のクエリ付与は未実装です（将来フェーズで対応予定）。

---

## safety

### safety.environment

| | |
|---|---|
| 型 | string |
| デフォルト | `staging` |
| 有効値 | `staging`, `production` |

`production` の場合、`--allow-production` フラグなしでは実行できません。

### safety.max_total_requests

| | |
|---|---|
| 型 | int |
| 必須 | いいえ |
| ハードリミット | `100,000` |

推定リクエスト数がこの値を超える場合、実行前にブロックされます。

### safety.allow_hosts

| | |
|---|---|
| 型 | string の配列 |
| 必須 | はい（最低 1 つ） |

ターゲット URL のホスト名がこのリストに含まれていない場合、実行は拒否されます。

---

## expect

### expect.block_status_codes

| | |
|---|---|
| 型 | int の配列 |
| 必須 | はい（最低 1 つ） |
| 例 | `[403, 429]` |

ブロック判定に使うステータスコード。

### expect.block_within_seconds

| | |
|---|---|
| 型 | int |
| 必須 | いいえ |

指定秒数以内にブロックステータスが出現することを期待。0 または未指定の場合、タイミング判定はスキップ。

### expect.min_block_ratio

| | |
|---|---|
| 型 | float (0.0〜1.0) |
| 必須 | いいえ |
| 例 | `0.20` (= 20%) |

全リクエスト中のブロック率がこの値以上であることを期待。0 または未指定の場合、比率判定はスキップ。

### expect.unaffected_paths

| | |
|---|---|
| 型 | 配列 |
| 必須 | いいえ |

誤検知チェック用。これらのパスがブロックされていないことを確認。

```yaml
expect:
  unaffected_paths:
    - path: /api/health
      method: GET
      expect_status: 200
```

> **注:** unaffected_paths の自動プローブは将来フェーズで実装予定です。

### expect.bypass_scenarios

| | |
|---|---|
| 型 | 配列 |
| 必須 | いいえ |

バイパスチェック用。特定のヘッダを付けたリクエストがブロックされないことを確認。

```yaml
expect:
  bypass_scenarios:
    - path: /assets/sample.jpg
      headers:
        X-Bypass-Token: valid-token
      expect_status: 200
```

> **注:** bypass_scenarios の自動プローブは将来フェーズで実装予定です。

---

## output

### output.json

| | |
|---|---|
| 型 | string |
| 必須 | いいえ |
| デフォルト | `results/result-<timestamp>-<name>.json` |

JSON 結果の出力パス。`{{timestamp}}` と `{{name}}` プレースホルダが使えます。

### output.markdown

| | |
|---|---|
| 型 | string |
| 必須 | いいえ |

Markdown レポートの出力パス。指定した場合、run 完了時に自動生成されます。

```yaml
output:
  json: results/result-{{timestamp}}-{{name}}.json
  markdown: results/report-{{timestamp}}-{{name}}.md
```
