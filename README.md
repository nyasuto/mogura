# Mogura

Mac のディスク使用量を詳細に分析する CLI ツール。

macOS 標準のストレージ管理では「システム」「その他」としか分からない情報を、ディレクトリ別に分解して可視化する。

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

### 出力例

```
Total: 12.3 GB (45678 files)

Path                          | Size
------------------------------+---------
/Users/ponpoko/Documents/proj | 3.2 GB
/Users/ponpoko/Documents/data | 2.1 GB
...
```

## 機能

- ディレクトリ再帰走査（シンボリックリンクはスキップ）
- ディレクトリ別サイズ集計（上位 10 件をテーブル表示）
- human-readable サイズ表示（B / KB / MB / GB / TB）
- パーミッションエラーは警告して続行

## 開発

```bash
make build     # ビルド
make test      # テスト
make quality   # vet + fmt + test
make clean     # バイナリ削除
make install   # go install
```

Ralph Loop（Claude Code 自律開発ループ）で開発：

```bash
./ralph.sh      # 自動ループ
./ralph.sh 10   # 最大10イテレーション
```

## ライセンス

MIT
