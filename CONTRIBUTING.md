# Contributing

## PR 前の品質チェック

PR を作成する前に、必ずローカルで `make quality` を実行してください。

```bash
make quality   # go vet + gofmt + go test
```

## CI 環境

GitHub Actions で以下の検証が自動実行されます:

- `go vet ./...`
- `gofmt` によるフォーマットチェック
- `go test -race ./...`（ローカルの `make test` に `-race` を追加）
- `go build ./...`
- `golangci-lint`（別ジョブ）

CI は **ubuntu-latest** と **macos-latest** の両方で実行されます。macOS は darwin 固有のビルドタグ（`getattrlistbulk` 等）の検証に必要です。

## `make quality` と CI の対応

| チェック項目 | `make quality` | CI |
|---|---|---|
| `go vet` | o | o |
| `gofmt` | o | o |
| `go test` | o | o (`-race` 付き) |
| `go build` | - | o |
| `golangci-lint` | `make lint` | o (別ジョブ) |

## コード規約

- 外部依存は原則ゼロ（標準ライブラリのみ）。`golang.org/x/sys` は例外
- テストはテーブルドリブンで書く
- エラーは明示的に返す（panic は使わない）
