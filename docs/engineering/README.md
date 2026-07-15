<!-- generated-by: scripts/generate_engineering_docs.py -->
# edge-checker — Engineering Handbook / Start Here

> 生成日: 2026-07-15 / 対象: `edge-checker` / 確度: [高]
> 実装・manifest・既存資料の静的棚卸しに基づく。外部サービスの稼働状態と本番構成は未検証。

## 60分で把握する

1. コンセプト: WAF/CDN 防御設定の検証ハーネス CLI ツール。 / 自社の WAF/CDN（Akamai 等）防御ルールが想定どおりに動作するかを、再現可能な形で検証し、pass/fail 判定とレポートを出力します。
2. classification: `active_project` / stack: Go, React ^19.2.4, TypeScript ~5.9.3, Tailwind CSS ^4.2.2, Node.js, Remotion, Vite
3. install: `pnpm install --frozen-lockfile`, `go mod download`
4. run/check: `npm run dev  # vite`, `npm run build  # tsc -b && vite build`, `npm run lint  # eslint .`
5. entrypoint: `cmd/edge-checker/main.go`, `web/src/main.tsx`

## 実装スナップショット

| 項目 | 現在値 | 最初に読むpath |
|---|---:|---|
| package/component | 5 | `web` |
| API | 0 | 未検出 |
| entity | 0 | 未検出 |
| screen/entry UI | 0 | 未検出 |
| test files | 12 | `internal/output/output_test.go` |

## 最初に確認する既存の正典候補

- `README.md`
- `architecture.drawio`
- `web/README.md`
- `docs/architecture.md`
- `docs/scenario-reference.md`
- `docs/usage-guide.md`

既存ADR、OpenAPI、schema、運用runbookがある場合は、下記generated docsより先に読む。

## 引継ぎblocking / partial

| Priority | Requirement | 状態・理由 | Evidence |
|---|---|---|---|
| P1 | `observability` | missing: 可観測性の実装証拠を特定できない。生成文書の一般要件・提案は現行実装の証拠ではない。 | `README.md` |
| P1 | `rollback` | missing: build/deploy可能だが実行可能なrollback手順がない。生成NFRもrelease前に定義としている。 | `docs/engineering/05_nfr_slo.md` |
| P1 | `screen_purpose_and_states` | missing: UI stackはあるがscreen/routeを特定できない。 | `architecture_diagram_notes.md` |
| P1 | `security` | partial: security control実装signalはあるが、運用契約・owner・検証手順が不足。 | `web/pnpm-lock.yaml` |

## 読む順番

1. [One Pager](./00_one_pager.md)
2. [技術スタック比較](./01_stack_comparison.md)
3. [アーキテクチャ・システム構成](./02_architecture.md)
4. [ADR](./03_adrs/ADR-0001-current-implementation-baseline.md)
5. [API定義](./04_api.md)
6. [データモデル・ER図](./05_data_model.md)
7. [非機能要件・SLO/SLI](./05_nfr_slo.md)
8. [画面設計](./06_screen_design.md)
9. [P50/P90見積り](./06_estimation.md)
10. [実装トレーサビリティ](./07_traceability.md)
11. [学習・保守ロードマップ](./08_learning_roadmap.md)

## 使い方

- generated docsは実装発見用handbook。既存ADR、OpenAPI、schema、runbookがある場合は既存正典を優先する。
- path・数・versionは静的検出した事実。目的やpath由来の責務は `[中]` の推定を含む。
- production、external console、secret値、migration適用状態は未確認。
