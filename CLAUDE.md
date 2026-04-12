# Mogura — Mac ディスク使用量アナライザ

> 指定ディレクトリ以下をスキャンし、ディレクトリ別・拡張子別・カテゴリ別のサイズ集計と不要ファイル検出を行う CLI ツール

## アーキテクチャ

```
CLI (flags) → Scanner (walk) → Analyzer (集計) → Formatter (出力)
```

## ディレクトリ構成

```
mogura/
├── main.go
└── internal/
    ├── types.go         # FileInfo 型、FormatSize
    ├── scanner/
    │   └── scanner.go   # ディレクトリ再帰走査
    ├── analyzer/
    │   └── directory.go # ディレクトリ別集計
    └── formatter/
        └── table.go     # テーブル出力
```

## ビルド・テストコマンド

```bash
go build ./...
go test ./...
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