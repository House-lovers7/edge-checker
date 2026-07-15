<!-- generated-by: scripts/generate_engineering_docs.py -->
# edge-checker — 技術スタック比較

> 生成日: 2026-07-15 / 対象: `edge-checker` / 確度: [高]
> 実装・manifest・既存資料の静的棚卸しに基づく。外部サービスの稼働状態と本番構成は未検証。

## 観測された採用スタック

Go, React ^19.2.4, TypeScript ~5.9.3, Tailwind CSS ^4.2.2, Node.js, Remotion, Vite

## Package / workspace

| Package | Manifest |
|---|---|
| `web` | `web/package.json` |

## トレードオフ

| 対象 | 現在 | 比較候補 | 現在案の利点 | 注意点 |
|---|---|---|---|---|
| Backend/CLI | Go | Python / Rust | single binary・標準toolchain | handler/data/UI検出はframework別確認が必要 |

> [中] 比較候補は現行実装を理解するための対照であり、移行提案ではない。当時の採用理由は既存ADRがあればそちらを正典とする。

## 判断を更新する条件

- manifest: `go.mod`, `web/package.json`
- quality config: `Makefile`, `web/tsconfig.json`, `web/eslint.config.js`, `.github/workflows/release.yml`, `.github/workflows/ci.yml`
- framework更新時はlockfile、build、typecheck、主要test、runtime smokeを同じ変更で確認する。
