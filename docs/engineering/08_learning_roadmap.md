<!-- generated-by: scripts/generate_engineering_docs.py -->
# edge-checker — 学習・保守ロードマップ

> 生成日: 2026-07-15 / 対象: `edge-checker` / 確度: [高]
> 実装・manifest・既存資料の静的棚卸しに基づく。外部サービスの稼働状態と本番構成は未検証。

## Day 1: 起動と全体像

1. install候補: `pnpm install --frozen-lockfile`, `go mod download`
2. 最初の実行/検査: `npm run dev  # vite`
3. `web` を読み、利用者向けUIとWebアプリケーション境界の境界を確認
4. `cmd/edge-checker` を読み、CLI・バッチ・運用入口の境界を確認
5. `internal` を読み、中核実装。詳細は配下moduleを参照の境界を確認
6. `cmd/edge-checker/main.go` を読み、実行entrypointの境界を確認
7. `web/src/main.tsx` を読み、実行entrypointの境界を確認

## Day 2–3: 主要契約

- APIがない/未検出であることを確認
- 永続化方式がfile/memory/external/なしのどれかを確定
- CLI/API/docs入口の成功・失敗フィードバックを確認
- external/config: 外部integration未検出 / 設定名未検出

## 最初の変更前

- 変更対象に最も近いtest: `internal/output/output_test.go`, `internal/cli/security_test.go`, `internal/observe/collector_test.go`, `internal/safety/policy_test.go`, `internal/safety/network_test.go`, `internal/safety/safety_test.go`, `internal/scenario/validator_test.go`, `internal/judge/probe_test.go`, `internal/judge/judge_test.go`, `internal/engine/burst_test.go`
- 既存ADR/docs: `README.md`, `architecture.drawio`, `web/README.md`, `docs/architecture.md`, `docs/scenario-reference.md`, `docs/usage-guide.md`
- runtime: Docker (`Dockerfile`), config (`.github/workflows/ci.yml`), config (`.github/workflows/release.yml`)
- `07_traceability.md` の未確認事項をcloseまたはrisk acceptしてから変更する。

## Doneの定義

- build/type/lint/testのうち存在するgateが通る。
- API/data/UI/runtimeの変更に対応する文書とADRを更新する。
- rollback、秘密情報、外部送信、production影響をreviewで明示する。
