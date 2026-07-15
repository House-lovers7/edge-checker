<!-- generated-by: scripts/generate_engineering_docs.py -->
# edge-checker — 実装トレーサビリティ

> 生成日: 2026-07-15 / 対象: `edge-checker` / 確度: [高]
> 実装・manifest・既存資料の静的棚卸しに基づく。外部サービスの稼働状態と本番構成は未検証。

## 根拠台帳

| 種別 | Path | 状態 |
|---|---|---|
| manifest | `go.mod` | 静的確認済み |
| manifest | `web/package.json` | 静的確認済み |
| entrypoint | `cmd/edge-checker/main.go` | 静的確認済み |
| entrypoint | `web/src/main.tsx` | 静的確認済み |
| package/source | `web/package.json` | 静的確認済み |
| package/source | `cmd/edge-checker` | 静的確認済み |
| package/source | `web` | 静的確認済み |
| package/source | `internal` | 静的確認済み |
| test | `internal/output/output_test.go` | 静的確認済み |
| test | `internal/cli/security_test.go` | 静的確認済み |
| test | `internal/observe/collector_test.go` | 静的確認済み |
| test | `internal/safety/policy_test.go` | 静的確認済み |
| test | `internal/safety/network_test.go` | 静的確認済み |
| test | `internal/safety/safety_test.go` | 静的確認済み |
| test | `internal/scenario/validator_test.go` | 静的確認済み |
| test | `internal/judge/probe_test.go` | 静的確認済み |
| test | `internal/judge/judge_test.go` | 静的確認済み |
| test | `internal/engine/burst_test.go` | 静的確認済み |
| test | `internal/engine/constant_test.go` | 静的確認済み |
| test | `internal/engine/ramp_test.go` | 静的確認済み |
| quality/CI | `Makefile` | 静的確認済み |
| quality/CI | `web/tsconfig.json` | 静的確認済み |
| quality/CI | `web/eslint.config.js` | 静的確認済み |
| quality/CI | `.github/workflows/release.yml` | 静的確認済み |
| quality/CI | `.github/workflows/ci.yml` | 静的確認済み |
| existing docs | `README.md` | 静的確認済み |
| existing docs | `architecture.drawio` | 静的確認済み |
| existing docs | `web/README.md` | 静的確認済み |
| existing docs | `docs/architecture.md` | 静的確認済み |
| existing docs | `docs/scenario-reference.md` | 静的確認済み |
| existing docs | `docs/usage-guide.md` | 静的確認済み |

## 検出した検証command

- **dev**: `npm run dev  # vite`
- **build**: `npm run build  # tsc -b && vite build`
- **lint**: `npm run lint  # eslint .`

## 設定契約（名前のみ）

- example/sourceから環境変数名を検出できず

値、credential、顧客データは収集していない。設定のrequired/optional、format、取得元は各entrypointとruntimeで確認する。

## 既存文書との関係

- `README.md`
- `architecture.drawio`
- `web/README.md`
- `docs/architecture.md`
- `docs/scenario-reference.md`
- `docs/usage-guide.md`

既存ADR・公式schema・運用runbookがある場合はそれらを正典とし、generated docsは発見用索引として扱う。矛盾を見つけたら実装・正式文書・生成器のどれを直すかをreviewで決める。

## 未確認事項

- 動的route/schema/plugin、external gateway、mobile native設定。
- secret manager、provider console、production runtimeの値と適用version。
- migration適用状態、SLO実績、実データ量、owner/on-call。

## 更新ルール

- route/schema/screen/runtime構成を変更した差分では、対応する文書を同時更新する。
- 生成し直す前に手書き文書を正典へ昇格するか、生成対象外へ分離する。
- このディレクトリの `generated-by` marker付きファイルは本スクリプトで再生成できる。
