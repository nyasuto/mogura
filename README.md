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

### オプション

```bash
# JSON 形式で出力
mogura --json ~/Documents

# ツリー形式で出力（深さ指定可能）
mogura --tree --depth 5 ~/Documents

# 巨大ファイル表示件数を変更（デフォルト: 20）
mogura --top 50 ~/Documents

# 古いファイルの判定日数を変更（デフォルト: 365）
mogura --older-than 180 ~/Documents

# 特定のディレクトリやファイルを除外
mogura --exclude 'node_modules,.git,*.tmp' ~/Projects

# HTML レポートを生成
mogura --html ~ > report.html
open report.html
```

### `-exclude` の使い方

`-exclude` フラグでスキャン対象から除外するパターンをカンマ区切りで指定できる。

| パターン種別 | 例 | マッチ対象 |
|---|---|---|
| ディレクトリ名の完全一致 | `node_modules` | パス中に `node_modules` という名前のディレクトリがあればスキップ |
| グロブ（ワイルドカード） | `*.tmp` | `filepath.Match` 準拠。ファイル名・ディレクトリ名に対してマッチ |

複数パターンの指定:

```bash
mogura --exclude 'node_modules,.git,.cache,*.log' ~/Projects
```

### 差分モード

`--json` で保存したスナップショットと現在の状態を比較し、ディレクトリごとの増減を確認できる。

```bash
# 1. 現在の状態をスナップショットとして保存
mogura --json ~ > snap_2026-04.json

# 2. 後日、前回のスナップショットと比較
mogura --diff snap_2026-04.json ~
```

テキスト出力では増加が赤、減少が緑で色分けされる。JSON / HTML 出力でも差分データが含まれ、HTML レポートではツリーマップが増減ヒートマップに切り替わる。

```bash
# JSON に差分情報を含めて出力
mogura --json --diff snap_2026-04.json ~ > snap_2026-07.json

# HTML ヒートマップで増減を可視化
mogura --html --diff snap_2026-04.json ~ > diff_report.html
open diff_report.html
```

### HTML レポート

`--html` フラグで、D3.js ツリーマップを使ったインタラクティブな HTML レポートを生成できる。

```bash
mogura --html ~ > report.html
open report.html
```

レポートには以下が含まれる:

- **ツリーマップ**: ディレクトリ構造をサイズ比例の矩形で可視化。カテゴリ別に色分けされる
- **ドリルダウン**: 矩形をクリックでサブディレクトリに移動。パンくずリストで階層を辿れる
- **ツールチップ**: ホバーでディレクトリ名・サイズ・ファイル数・全体割合を表示
- **サマリパネル**: 総容量・カテゴリ内訳の棒グラフ・キャッシュ合計・推定節約可能量
- **巨大ファイル Top10**: サイズの大きいファイルをテーブルで一覧表示

HTML ファイルは単体で動作する（D3.js は CDN から読み込み）。

## 機能

- ディレクトリ再帰走査（シンボリックリンクはスキップ）
- ディレクトリ別サイズ集計（上位 10 件をテーブル表示）
- 拡張子別・カテゴリ別サイズ集計
- 巨大ファイル Top N 表示
- キャッシュ・ビルド成果物などのゴミディレクトリ検出
- 古いファイル（未更新 N 日以上）の検出と節約可能量の推定
- human-readable サイズ表示（B / KB / MB / GB / TB）
- JSON / ツリー / テーブル / HTML 出力形式の切り替え
- HTML レポートによるインタラクティブなツリーマップ可視化
- 差分モードによる前回スナップショットとの増減比較
- 除外パターンによるスキャン対象の絞り込み
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
