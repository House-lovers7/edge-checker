# 実務ワークフローガイド

edge-checker を使った WAF/CDN 防御設定の検証ワークフローを解説します。

---

## 基本フロー

どの検証でも共通のフローは以下の 5 ステップです。

```
1. シナリオ作成  →  2. validate  →  3. dry-run  →  4. run  →  5. 報告
```

### ステップ 1: シナリオ作成

`scenarios/` にシナリオ YAML を作成します。既存のサンプルをコピーして編集するのが一番早いです。

```bash
cp scenarios/assets-threshold.yaml scenarios/my-check.yaml
# → base_url, path, rps, expect 等を自分の環境に合わせて編集
```

### ステップ 2: validate

```bash
./edge-checker validate -f scenarios/my-check.yaml
```

YAML の構文エラー、必須フィールドの欠落、設定の矛盾を検出します。

### ステップ 3: dry-run

```bash
./edge-checker dry-run -f scenarios/my-check.yaml
```

確認すべき点:
- ターゲット URL は正しいか
- 推定リクエスト数は想定通りか
- 安全チェックが全て OK か
- 判定基準は意図通りか

### ステップ 4: run

```bash
./edge-checker run -f scenarios/my-check.yaml
```

実行中はプログレスバーが表示されます:
```
[==========          ] 15/30s | 900 reqs | 200:720 403:180 | p95: 42ms
```

完了後、自動で判定結果が表示されます。

### ステップ 5: 報告

```bash
./edge-checker report -r results/result-XXXXXXXX-XXXXXX-my-check.json -o reports/my-check.md
```

生成された Markdown をそのまま社内チケットや Confluence に貼ります。

---

## ワークフロー別ガイド

### A. Rate Policy のしきい値検証

**目的:** Rate Policy が設定した rps で正しく発火するかを確認する

#### A-1: しきい値以下（発火しないことを確認）

```yaml
name: threshold-below
description: Confirm no blocking at 50% of threshold

target:
  base_url: https://staging.example.com
  path: /assets/images/medium/sample.jpg

execution:
  mode: constant
  duration: 30s
  concurrency: 3

rate:
  rps: 30    # しきい値の 50% に設定

profile: browser-like

safety:
  allow_hosts: [staging.example.com]

expect:
  block_status_codes: [403, 429]
  min_block_ratio: 0.0    # ブロックが 0% であることを期待
```

> **注:** 現在の判定ロジックでは「ブロック率が 0% であること」の判定は未実装です。
> ブロックステータスが出現しなければ Block detection が FAIL になるので、この場合は FAIL = 期待通りです。

#### A-2: しきい値超過（発火することを確認）

```yaml
name: threshold-above
description: Confirm blocking triggers at 150% of threshold

execution:
  mode: constant
  duration: 30s
  concurrency: 5

rate:
  rps: 90    # しきい値の 150% に設定

expect:
  block_status_codes: [403, 429]
  block_within_seconds: 30
  min_block_ratio: 0.20
```

#### A-3: 結果の比較

```bash
# 実行
./edge-checker run -f scenarios/threshold-below.yaml -o results/below.json
./edge-checker run -f scenarios/threshold-above.yaml -o results/above.json

# 比較
./edge-checker show -r results/below.json
./edge-checker show -r results/above.json
```

期待する結果:
- below: ブロックなし（全て 200）
- above: 一定割合がブロックされる（403/429 が出現）

---

### B. バースト耐性の検証

**目的:** 短時間の急増アクセスで Rate Policy が発火するかを確認する

```yaml
name: burst-check
description: Verify burst detection triggers correctly

execution:
  mode: burst
  duration: 60s
  concurrency: 10

rate:
  rps: 15
  burst:
    spike_rps: 100
    spike_duration: 5s
    interval: 15s

profile: bot-like

expect:
  block_status_codes: [403, 429]
  block_within_seconds: 30
  min_block_ratio: 0.15
```

**確認ポイント:**
- 最初のスパイクで発火するか、2 回目以降か
- スパイク中だけブロックされるか、スパイク後もブロックが続くか（Penalty Box）
- `show` コマンドのタイムライン出力でブロック開始タイミングを確認

---

### C. Penalty Box 解除タイミングの確認

**目的:** ブロック後、何分で解除されるかを確認する

```yaml
name: penalty-box-check
description: Measure Penalty Box release timing

execution:
  mode: cooldown-check
  duration: 120s
  concurrency: 5

rate:
  rps: 60
  cooldown:
    active_duration: 15s    # 15秒間高レートで発火させる
    active_rps: 80
    wait_duration: 60s      # 60秒間プローブで待機
    probe_rps: 1

expect:
  block_status_codes: [403, 429]
  block_within_seconds: 15
  min_block_ratio: 0.10
```

**実行時の出力:**
```
  [Phase 1/3] Active: 80 rps for 15s
  [Phase 2/3] Cooldown: probing at 1 rps for 60s
  [Phase 3/3] Verify: 40 rps for 5s
```

**結果の見方:**
- タイムラインの Phase 2 区間で、403/429 → 200 に切り替わるタイミングが Penalty Box の解除時刻
- Phase 3 で 200 が返れば復帰確認完了

---

### D. Bypass IP の差分検証

**目的:** Bypass IP は本当に除外されているか、非 Bypass IP はちゃんと検知されるかを確認する

同じシナリオを 2 つの環境から実行して結果を比較します。

```yaml
# scenarios/bypass-compare.yaml（共通シナリオ）
name: bypass-comparison
description: Compare bypass vs non-bypass behavior

target:
  base_url: https://staging.example.com
  path: /assets/images/medium/sample.jpg

execution:
  mode: constant
  duration: 30s
  concurrency: 5

rate:
  rps: 60

profile: browser-like

safety:
  allow_hosts: [staging.example.com]

expect:
  block_status_codes: [403, 429]
  block_within_seconds: 30
  min_block_ratio: 0.20
```

**実行手順:**

```bash
# 1. 非 Bypass IP のマシンから実行（ブロックされるべき）
./edge-checker run -f scenarios/bypass-compare.yaml -o results/non-bypass.json

# 2. Bypass IP のマシンから実行（ブロックされないべき）
./edge-checker run -f scenarios/bypass-compare.yaml -o results/bypass.json

# 3. 結果比較
./edge-checker show -r results/non-bypass.json
./edge-checker show -r results/bypass.json
```

**期待する差分:**
- non-bypass: PASS（ブロックされた = 防御が効いている）
- bypass: FAIL（ブロックされなかった = bypass が効いている）

> **重要:** bypass 側が PASS（ブロックされた）になる場合、Bypass 設定に問題があります。

---

### E. しきい値の境界探索

**目的:** 「何 rps から引っかかるか」を特定する

```yaml
name: ramp-boundary
description: Find the exact rate policy trigger point

execution:
  mode: ramp
  duration: 60s
  concurrency: 10

rate:
  rps: 10
  ramp:
    start_rps: 5
    end_rps: 100
    step_duration: 10s

expect:
  block_status_codes: [403, 429]
  block_within_seconds: 60
  min_block_ratio: 0.10
```

**結果の見方:**
- JSON のタイムラインで、最初に 403/429 が出現する秒数を確認
- その秒数での RPS を逆算すれば、しきい値の近似値がわかる

```bash
./edge-checker run -f scenarios/ramp-boundary.yaml
./edge-checker show -r results/result-XXXXXXXX-ramp-boundary.json
```

例: 30 秒後に 403 が出始めた場合
- 10 秒ごとのステップで 5 → 100 に増加
- 30 秒後は step 3 = 5 + (100-5)/6 * 3 ≈ 52 rps
- しきい値はおよそ 50 rps 付近

---

### F. 設定変更の回帰テスト

**目的:** Rate Policy を変更した後に、変更が正しく反映されたかを確認する

**手順:**

1. 変更前に現在の動作を記録:
```bash
./edge-checker run -f scenarios/before-change.yaml -o results/before.json
./edge-checker report -r results/before.json -o reports/before.md
```

2. Rate Policy を変更

3. 変更後に同じシナリオで再テスト:
```bash
./edge-checker run -f scenarios/before-change.yaml -o results/after.json
./edge-checker report -r results/after.json -o reports/after.md
```

4. 結果を比較して変更の効果を確認

---

## 複数パスの検証

対象パスが複数ある場合、パスごとにシナリオを作成して順次実行します。

```bash
# /assets/images/medium/* の検証
./edge-checker run -f scenarios/images-threshold.yaml

# /assets/materials/* の検証
./edge-checker run -f scenarios/materials-threshold.yaml

# 対象外パス（誤爆しないことの確認）
./edge-checker run -f scenarios/html-pages-check.yaml
```

---

## レポートの活用

### 社内報告

```bash
./edge-checker report -r results/result.json -o reports/rate-policy-test-20260326.md
```

生成された Markdown には以下が含まれます:
- シナリオ概要（何を、いつ、どうテストしたか）
- サマリ（リクエスト数、ステータス分布、レイテンシ）
- 判定結果（PASS/FAIL テーブル）
- 秒単位タイムライン

### 証跡としての保存

JSON 結果は Git で管理するか、チケットに添付することで、以下に使えます:
- 設定変更の事前・事後比較の証跡
- 障害調査時の「当時の設定は効いていたか」の確認
- 引き継ぎ用の標準テスト手順と実績

---

## Tips

### RPS の見積もり

Rate Policy のしきい値が不明な場合:
1. `ramp` モードで境界を探索
2. 見つかった境界付近で `constant` モードで再確認
3. 境界の 50%, 80%, 100%, 150% でそれぞれ確認

### 並列数 (concurrency) の決め方

- しきい値検証: `3〜5` で十分
- バースト検証: `10〜20`（スパイク時に詰まらない程度）
- 負荷が軽い場合: `1` でも OK（RPS がボトルネックになるので）

### タイムアウトの設定

- 通常: `10s`（デフォルト）
- CDN 経由でレイテンシが高い場合: `15s〜30s`
- ブロック時に応答が遅い場合: `timeout` を長めにしないとエラーにカウントされる
