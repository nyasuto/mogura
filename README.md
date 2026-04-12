# Mogura 🐹

Mac のディスク使用量を詳細に分析する CLI ツール。

macOS 標準のストレージ管理では「システム 200GB」「その他 150GB」としか分からない情報を、ディレクトリ別・拡張子別・カテゴリ別に分解して可視化する。

## インストール

```bash
go install github.com/nyasuto/mogura@latest
```

## 使い方

```bash
# ホームディレクトリを分析
mogura ~

# 特定ディレクトリを分析
mogura /Users/ponpoko/Documents
```

## 開発

Ralph Loop（Claude Code 自律開発ループ）で開発：

```bash
./ralph.sh      # 自動ループ
./ralph.sh 10   # 最大10イテレーション
```

## ライセンス

MIT