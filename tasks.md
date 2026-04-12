# Tasks

## Phase 1: スキャン基盤

- [x] プロジェクト初期化（go.mod, main.go, internal/scanner/, internal/analyzer/, internal/formatter/ ディレクトリ作成）
- [x] internal/types.go — FileInfo 型定義 + FormatSize 関数（human-readable サイズ変換）+ テスト
- [x] internal/scanner/scanner.go — ディレクトリ再帰走査。シンボリックリンクはスキップ、パーミッションエラーは stderr 警告して継続。FileInfo のスライスを返す + テスト
- [x] internal/analyzer/directory.go — ディレクトリ別サイズ集計。map[string]int64 を返す + テスト
- [x] main.go — CLI エントリポイント。引数でパスを受け取り、合計サイズとディレクトリ上位10件をテーブル表示
- [x] makefileのような開発エコシステムを作る
- [x] Readme, CLAUDE.md を最新コードベースを元に更新

## Phase 2: 分析機能

- [x] internal/analyzer/extension.go — 拡張子別サイズ集計。拡張子ごとの合計サイズとファイル数を返す構造体を定義 + テスト
- [x] internal/analyzer/category.go — カテゴリ分類ルール定義。拡張子 → カテゴリ（動画/画像/コード/ドキュメント/キャッシュ/アーカイブ/その他）のマッピングを map で定義
- [x] internal/analyzer/category.go — カテゴリ別サイズ集計。カテゴリごとの合計サイズ・ファイル数・全体に対する割合を返す + テスト
- [x] internal/analyzer/topn.go — 巨大ファイル Top N。サイズ降順で上位 N 件の FileInfo を返す（デフォルト N=20）+ テスト
- [x] main.go — Phase 2 の分析結果を CLI 出力に統合。拡張子 Top10、カテゴリ内訳、巨大ファイル Top20 をセクション分けして表示

## Phase 3: 出力フォーマット

### 3-A: ディレクトリツリー構造

- [x] internal/analyzer/tree.go — DirNode 型の定義のみ。Name string, Size int64, Children []DirNode, FileCount int のフィールドを持つ
- [x] internal/analyzer/tree.go — BuildTree 関数。FileInfo スライスからルート DirNode を構築する。全階層を保持する + テスト
- [x] internal/analyzer/tree.go — Prune 関数。DirNode を指定 depth で刈り込み、それより深い階層のサイズは親に集約する + テスト

### 3-B: テーブルフォーマッタ

- [x] internal/formatter/table.go — Row 型（[]string）と Table 型（Header + Rows）を定義。Render 関数で列幅を自動計算し、サイズ列は右寄せで整形して文字列を返す + テスト
- [x] main.go — Phase 1, 2 の既存テーブル出力を formatter/table.go に置き換えるリファクタリング。出力結果は変えない

### 3-C: ツリーフォーマッタ

- [x] internal/formatter/tree.go — RenderTree 関数。DirNode を tree コマンド風のインデント付きテキストに変換。各行に「名前 サイズ (割合%)」を表示する + テスト
- [x] internal/formatter/tree.go — 割合計算の追加。各 DirNode のサイズをルートサイズに対するパーセンテージで表示。1% 未満のノードは省略する + テスト

### 3-D: JSON フォーマッタ

- [x] internal/formatter/json.go — Report 型を定義。TotalSize, ScannedAt, DirTree(DirNode), Extensions(拡張子集計), Categories(カテゴリ集計), LargestFiles(Top N) を全て含む単一構造体
- [x] internal/formatter/json.go — RenderJSON 関数。Report を encoding/json で整形出力（indent 付き）。DirNode のネスト構造はそのまま保持する（GUI ツリーマップ用） + テスト

### 3-E: CLI フラグ統合

- [x] main.go — flag パッケージで CLI フラグを追加。`--json`, `--tree`, `--depth N`(デフォルト3), `--top N`(デフォルト20)
- [x] main.go — フラグに応じた出力分岐。`--json` → JSON 全量出力、`--tree` → ツリー表示、フラグなし → テーブル表示（現行動作）

## Phase 4: ゴミ発見

### 4-A: 既知ディレクトリの検出

- [x] internal/analyzer/waste.go — WasteDir 型の定義。Path, Size, Kind(string) のフィールド。Kind は "node_modules", "cache", "git", "build" 等の分類ラベル
- [x] internal/analyzer/waste.go — 検出対象パターンのリスト定義。node_modules, .cache, __pycache__, DerivedData, .Trash, Caches(~/Library/Caches), .gradle, .cargo/registry, .npm, target(Rust) 等を map[string]string（ディレクトリ名 → Kind）で定義
- [x] internal/analyzer/waste.go — DetectWaste 関数。FileInfo スライスからパターンに一致するディレクトリを検出し、WasteDir のスライスを返す。サイズ降順でソート + テスト
- [x] internal/analyzer/waste.go — DetectLargeGitDirs 関数。.git ディレクトリのうちサイズが閾値（デフォルト 100MB）以上のものを検出 + テスト

### 4-B: 古いファイルの検出

- [x] internal/analyzer/stale.go — DetectStale 関数。最終更新が N 日以上前のファイルを検出し、合計サイズとファイル数を返す。ディレクトリ別にグルーピング + テスト

### 4-C: サマリレポート

- [ ] internal/formatter/summary.go — サマリレポートの生成。総容量、カテゴリ内訳上位5件、キャッシュ合計サイズ、古いファイル合計サイズ、推定節約可能量をコンパクトに表示 + テスト
- [ ] internal/formatter/json.go — Report 型に WasteDirs, StaleSummary, SavingsEstimate フィールドを追加。JSON 出力に反映
- [ ] main.go — CLI にゴミ発見結果を統合。`--older-than N`（日数、デフォルト 365）フラグを追加。デフォルト出力にサマリセクションを追加

---

## Backlog

---

## 設計メモ