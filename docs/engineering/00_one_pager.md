<!-- generated-by: scripts/generate_engineering_docs.py -->
# edge-checker — One Pager / オンボーディング概要

> 生成日: 2026-07-15 / 対象: `edge-checker` / 確度: [高]
> 実装・manifest・既存資料の静的棚卸しに基づく。外部サービスの稼働状態と本番構成は未検証。

## コンセプト

WAF/CDN 防御設定の検証ハーネス CLI ツール。 / 自社の WAF/CDN（Akamai 等）防御ルールが想定どおりに動作するかを、再現可能な形で検証し、pass/fail 判定とレポートを出力します。

## 誰の何を解くか

- 対象領域: セキュリティ/認証/監査/決済
- 想定利用者: 開発/セキュリティ/運用担当
- 価値仮説: 設定/ログ/リクエスト/差分を検査し、危険差分・承認・監査証跡・決済境界を可視化。

## 現在地

| 項目 | 観測結果 |
|---|---|
| 技術スタック | Go, React, TypeScript, Tailwind CSS, Node.js, Remotion, Vite |
| API | 0 endpoint signal |
| データモデル | 0 unique entity signal |
| 画面 | 0 route/screen signal |
| 実行基盤 | Docker (`Dockerfile`), config (`.github/workflows/ci.yml`), config (`.github/workflows/release.yml`) |
| package / module | 5 component signal |
| tests | 12 file signal |

## ソースマップ

| Component | Path | 責務 |
|---|---|---|
| `web` | `web` | 利用者向けUIとWebアプリケーション境界 |
| `edge-checker` | `cmd/edge-checker` | CLI・バッチ・運用入口 |
| `internal` | `internal` | 中核実装。詳細は配下moduleを参照 |
| `main` | `cmd/edge-checker/main.go` | 実行entrypoint |
| `main` | `web/src/main.tsx` | 実行entrypoint |

## 最初に使うコマンド

| 目的 | Command |
|---|---|
| `dev` | `npm run dev  # vite` |
| `build` | `npm run build  # tsc -b && vite build` |
| `lint` | `npm run lint  # eslint .` |

## 変更箇所の入口

| 変更対象 | 最初に読むpath | 同時に確認するもの |
|---|---|---|
| 実行・配備 | `Dockerfile` | 環境変数、service依存、rollback |
| 回帰検査 | `internal/output/output_test.go` | 変更対象に近いtestと全体check |

## 引継ぎ時の未解決ギャップ

| Priority | Requirement | 状態・理由 | Evidence |
|---|---|---|---|
| P1 | `observability` | missing: 可観測性の実装証拠を特定できない。生成文書の一般要件・提案は現行実装の証拠ではない。 | `README.md` |
| P1 | `rollback` | missing: build/deploy可能だが実行可能なrollback手順がない。生成NFRもrelease前に定義としている。 | `docs/engineering/05_nfr_slo.md` |
| P1 | `screen_purpose_and_states` | missing: UI stackはあるがscreen/routeを特定できない。 | `architecture_diagram_notes.md` |
| P1 | `security` | partial: security control実装signalはあるが、運用契約・owner・検証手順が不足。 | `web/pnpm-lock.yaml` |

## スコープ境界

- [高] productionの稼働、外部provider設定、secret値は未確認。
- [高] API・DB・画面が未検出の場合は推測せず、実装入口の追加を課題として残す。
- [中] 初回変更前に `07_traceability.md` の根拠と未確認事項を確認する。
