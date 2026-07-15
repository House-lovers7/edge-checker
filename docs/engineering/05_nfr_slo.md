<!-- generated-by: scripts/generate_engineering_docs.py -->
# edge-checker — 非機能要件・SLO/SLI

> 生成日: 2026-07-15 / 対象: `edge-checker` / 確度: [高]
> 実装・manifest・既存資料の静的棚卸しに基づく。外部サービスの稼働状態と本番構成は未検証。

## 現在コード化されている品質ゲート

| Gate | Command | 根拠 |
|---|---|---|
| dev | `npm run dev  # vite` | manifest script |
| build | `npm run build  # tsc -b && vite build` | manifest script |
| lint | `npm run lint  # eslint .` | manifest script |

- test files: 12（`internal/output/output_test.go`, `internal/cli/security_test.go`, `internal/observe/collector_test.go`, `internal/safety/policy_test.go`, `internal/safety/network_test.go`, `internal/safety/safety_test.go`, `internal/scenario/validator_test.go`, `internal/judge/probe_test.go`, `internal/judge/judge_test.go`, `internal/engine/burst_test.go`, `internal/engine/constant_test.go`, `internal/engine/ramp_test.go`）
- quality/CI config: `Makefile`, `web/tsconfig.json`, `web/eslint.config.js`, `.github/workflows/release.yml`, `.github/workflows/ci.yml`
- security/resilience signal: rate-limit (`web/src/utils/packetSampler.ts`), rate-limit (`web/src/components/dashboard/SummaryCards.tsx`), rate-limit (`web/src/components/flow/Packet.tsx`), resilience (`internal/httpclient/client.go`), resilience (`internal/cli/security_test.go`), resilience (`internal/cli/run.go`), auth/session (`internal/cli/dryrun.go`), resilience (`internal/safety/dns.go`), auth/session (`internal/safety/policy_test.go`), auth/session (`internal/safety/policy.go`), auth/session (`internal/safety/preflight.go`), resilience (`internal/scenario/validator.go`), resilience (`internal/scenario/loader.go`), resilience (`internal/scenario/validator_test.go`), resilience (`internal/scenario/model.go`), rate-limit (`internal/engine/ramp.go`), rate-limit (`internal/engine/cooldown.go`), rate-limit (`internal/engine/burst.go`), rate-limit (`internal/engine/constant.go`)

## 計測すべきSLI

| Boundary | SLI | 最初の計測根拠 |
|---|---|---|
| CLI/Job | exit code・処理件数・失敗件数・処理時間 | `cmd/edge-checker/main.go` |

## SLOの状態

[高] 合意済みSLO数値はrepository内の実装・資料から確認できていない。任意の99%や2秒を現在要件として記載しない。利用者、運用時間帯、障害コスト、予算を確認してから、上記SLIごとにtarget/window/error budgetを決める。

## 運用境界

- runtime/config: Docker (`Dockerfile`), config (`.github/workflows/ci.yml`), config (`.github/workflows/release.yml`)
- required config names: example/sourceから未検出
- 外部integration: 静的検出なし
- rollbackはcode、schema、generated artifact、provider設定を分ける。production操作は人間承認後に行う。
