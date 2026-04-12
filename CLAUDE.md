# Mogura — Mac ディスク使用量アナライザ

> 指定ディレクトリ以下をスキャンし、ディレクトリ別のサイズ集計を行う CLI ツール

## アーキテクチャ

```
CLI (args) → Scanner (walk) → Analyzer (集計) → Formatter (出力)
```

## ディレクトリ構成

```
.
├── main.go                          # エントリポイント: 引数パス → scan → 集計 → 表示
├── Makefile                         # build / test / vet / fmt / quality / clean / install
├── go.mod                           # module mogura, Go 1.26.2
└── internal/
    ├── types.go                     # FileInfo 型、FormatSize（human-readable 変換）
    ├── types_test.go
    ├── scanner/
    │   ├── scanner.go               # ディレクトリ再帰走査（symlink スキップ、permission 警告続行）
    │   └── scanner_test.go
    ├── analyzer/
    │   ├── directory.go             # ディレクトリ別サイズ集計（map[string]int64）
    │   └── directory_test.go
    └── formatter/
        └── table.go                 # サイズ降順ソート → 上位 N 件テーブル出力
```

## ビルド・テストコマンド

```bash
make build     # go build -o mogura .
make test      # go test ./...
make quality   # vet + fmt + test
go vet ./...
```

## コード規約

- 外部依存ゼロ（標準ライブラリのみ）
- テストはテーブルドリブンで書く
- エラーは明示的に返す（panic は使わない）
- サイズは内部で常に int64（バイト）、表示時のみ human-readable 変換
- シンボリックリンクはスキップする
- パーミッションエラーは stderr に警告して続行する

## パッケージ依存ルール

types ← scanner ← analyzer ← formatter（逆方向の依存は禁止）

## やってはいけないこと

- 外部ライブラリへの依存追加
- ファイルの削除機能の実装（分析と表示のみ）
- tasks.md に記載されていないタスクの着手
