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

- [x] internal/formatter/summary.go — サマリレポートの生成。総容量、カテゴリ内訳上位5件、キャッシュ合計サイズ、古いファイル合計サイズ、推定節約可能量をコンパクトに表示 + テスト
- [x] internal/formatter/json.go — Report 型に WasteDirs, StaleSummary, SavingsEstimate フィールドを追加。JSON 出力に反映
- [x] main.go — CLI にゴミ発見結果を統合。`--older-than N`（日数、デフォルト 365）フラグを追加。デフォルト出力にサマリセクションを追加


## Phase 5: 土台整理 + 実用性
 
### 5-A: main.go のオーケストレーション層抽出（R1）
 
- [x] internal/app/app.go — Config 型の定義。TargetPath, TopN, Depth, OutputFormat(text/json/tree/html), Exclude[]string, OlderThanDays のフィールドを持つ
- [x] internal/app/app.go — Run(cfg Config, stdout io.Writer, stderr io.Writer) error 関数のシグネチャと空実装。main.go から呼べることを確認
- [x] main.go → internal/app/app.go — flag 解析ロジックを ParseFlags(args []string) (Config, error) として移動。main.go は ParseFlags → Run だけにする
- [x] internal/app/app.go — Run 関数にスキャン → 分析 → 出力のロジックを main.go から移植。main.go は 20 行以内になること
- [x] internal/app/app_test.go — Run 関数の基本テスト。テスト用 tmpdir をスキャンして出力が空でないことを確認
 
### 5-B: Result 集約型の導入（R2）
 
- [x] internal/analyzer/result.go — Result 型の定義。TotalSize, DirSizes, ExtStats, CategoryStats, TopFiles, DirTree, WasteDirs, StaleSummary を全て持つ構造体
- [x] internal/analyzer/analyze.go — Analyze(files []types.FileInfo, opts AnalyzeOpts) Result 関数。既存の各 analyzer を内部で呼び出して Result に詰める + テスト
- [x] internal/app/app.go — Run 関数を Analyze 呼び出しに書き換え。個別の analyzer 呼び出しを削除
- [x] internal/formatter/ — 各フォーマッタを Result を受け取る形に統一。FormatTable(Result, io.Writer), FormatJSON(Result, io.Writer), FormatTree(Result, io.Writer)
- [x] 旧 main.go 内の analyzer 個別呼び出しコードが完全に消えていることを確認。go vet + go test パス
 
### 5-C: 除外パターン（F2）
 
- [x] internal/scanner/scanner.go — ScanOpts 型に Exclude []string フィールドを追加。Walk 時にパスがパターンに一致したらスキップ + テスト
- [x] パターンマッチの仕様決定: ディレクトリ名の完全一致（`node_modules`）とグロブ（`*.tmp`）の両方を対応 + テスト
- [x] internal/app/app.go — Config.Exclude を ScanOpts に渡す配線。`-exclude 'node_modules,.git,*.tmp'` をカンマ区切りでパース
- [x] README.md に `-exclude` の使い方を追記
 
## Phase 6: HTML ツリーマップ
 
### 6-A: HTML テンプレート基盤
 
- [x] internal/formatter/html.go — FormatHTML(Result, io.Writer) error のシグネチャと空実装
- [x] internal/formatter/templates/ ディレクトリ作成。report.html テンプレートの骨格を作成（html/template 用）。D3.js は cdnjs から読み込む script タグ
- [x] internal/formatter/html.go — go:embed で templates/report.html を埋め込み。Result の JSON を `<script>const data = {{.}}</script>` としてテンプレートに注入
- [x] internal/formatter/html.go — テスト。Result のモックデータで HTML を生成し、`const data =` が含まれることを確認
 
### 6-B: ツリーマップ描画
 
- [x] templates/report.html — D3.js treemap レイアウトの実装。DirTree の children を d3.hierarchy → d3.treemap で矩形配置。各矩形にディレクトリ名とサイズを表示
- [x] templates/report.html — カテゴリ別の色分け。ディレクトリ内の支配的なカテゴリ（動画/画像/コード等）で矩形の背景色を変える
- [x] templates/report.html — クリックでドリルダウン。矩形をクリックするとそのディレクトリを新しいルートとしてツリーマップを再描画。パンくずリストで階層を表示
- [x] templates/report.html — ホバー時にツールチップ。ディレクトリ名、サイズ（human-readable）、ファイル数、全体に対する割合を表示
 
### 6-C: サマリパネル
 
- [x] templates/report.html — ツリーマップの上にサマリパネルを追加。総容量、カテゴリ内訳（円グラフ or 棒グラフ）、キャッシュ合計、推定節約可能量
- [x] templates/report.html — 巨大ファイル Top10 のテーブルをサイドパネルまたは下部に表示
 
### 6-D: CLI 統合
 
- [x] internal/app/app.go — OutputFormat が html のとき FormatHTML を呼び出す分岐を追加
- [x] main.go / flags — `-html` フラグの追加。stdout に HTML を出力（`mogura --html ~ > report.html`）
- [x] README.md に HTML レポートの使い方とスクリーンショット説明を追記
 
## Phase 7: 差分モード
 
### 7-A: 差分計算
 
- [x] internal/analyzer/diff.go — DirDiff 型の定義。Path, PrevSize, CurrSize, Delta int64 のフィールド
- [x] internal/analyzer/diff.go — LoadPrevResult(path string) (Result, error) 関数。JSON ファイルを読み込んで Result にデコード + テスト
- [x] internal/analyzer/diff.go — ComputeDiff(prev, curr Result) []DirDiff 関数。ディレクトリ別のサイズ差分を計算。新規ディレクトリと削除ディレクトリも検出。Delta 降順でソート + テスト
 
### 7-B: 差分出力
 
- [x] internal/formatter/table.go — FormatDiffTable 関数。増減を +/- 付きで表示、増加は赤系・減少は緑系の ANSI カラー + テスト
- [x] internal/formatter/json.go — Result に DiffSummary フィールドを追加（optional）。差分モード時のみ値が入る
- [x] templates/report.html — 差分モード時にツリーマップの矩形色を「増加=赤、減少=青、変化なし=グレー」のヒートマップに切り替え
 
### 7-C: CLI 統合
 
- [x] internal/app/app.go — Config に DiffPath string を追加。`-diff prev.json` で前回 JSON を読み込み ComputeDiff を呼ぶ
- [x] README.md に差分モードのワークフロー（`mogura --json ~ > snap.json` → 後日 `mogura --diff snap.json ~`）を追記
 
## Phase 8: CLI 磨き込み
 
### 8-A: ASCII 棒グラフ（V1）
 
- [x] internal/formatter/bar.go — RenderBar(value, maxValue, width int) string 関数。`████████░░░░` を返す + テスト
- [x] internal/formatter/table.go — カテゴリ・拡張子・ディレクトリのテーブル行末に RenderBar を追加
 
### 8-B: フィルタ（F7）
 
- [x] internal/app/app.go — Config に MinSize int64, FilterExt []string を追加
- [x] internal/app/app.go — スキャン結果に対して MinSize / FilterExt でフィルタリングする関数。Analyze の前段で適用 + テスト
- [x] flags — `-min-size 10M`（human-readable パース）、`-ext mp4,mkv` フラグを追加
 
### 8-C: 進捗表示（F6）
 
- [x] internal/scanner/scanner.go — ScanOpts に OnProgress func(scanned int, currentDir string) コールバックを追加
- [x] internal/app/app.go — stderr に `Scanning... 12345 files (~/Library/Caches)` を 500ms 間隔で表示。`-quiet` フラグで抑制
 
### 8-D: マウント境界（F12）
 
- [x] internal/scanner/scanner.go — ScanOpts に OneFileSystem bool を追加。Walk 時にデバイス ID が変わったらスキップ（syscall.Stat_t.Dev を比較） + テスト
- [x] flags — `-x` フラグの追加


## Phase 9: スパースファイル対応（F0 — 実証済みの最重要課題）

> 実スキャンで `~/.colima/_lima/.../diffdisk` が論理 100 GB / 実物理 9.2 GB、`Docker.raw` も論理 926 GB で同様の乖離が判明。現状の mogura は `stat.Size` しか見ていないため「推定節約可能量」が実態と大幅に乖離する。`stat.Blocks * 512` で物理サイズを取得し、両軸で可視化する。

### 9-A: 物理サイズの取得

- [x] internal/types.go — FileInfo に PhysicalSize int64 フィールドを追加。JSON タグは `physical_size`（既存 `size` は論理サイズのまま据え置き = 後方互換）
- [x] internal/scanner/scanner.go — darwin / linux 用に syscall.Stat_t.Blocks を読み取って PhysicalSize に詰める実装。既存の Lstat 取得経路に統合（重複 stat は避ける）+ テスト
- [x] internal/scanner/scanner_test.go — `os.Truncate` で巨大論理サイズのスパースファイルを一時生成し、PhysicalSize << Size になることを検証するテストケース

### 9-B: 集計への反映

- [x] internal/analyzer/result.go — Result に TotalPhysicalSize int64 を追加
- [x] internal/analyzer/directory.go — AggregateByDir を `map[string]struct{Size,Physical int64}` に変更（or 並列 map を追加）+ テスト更新
- [x] internal/analyzer/extension.go / category.go — ExtStats / CategoryStats に PhysicalSize を追加 + テスト更新
- [x] internal/analyzer/topn.go — 巨大ファイル Top は論理サイズで並べたうえで物理サイズも保持（並び順は現行維持）
- [x] internal/analyzer/tree.go — DirNode に PhysicalSize を追加し BuildTree / Prune で積み上げ + テスト
- [x] internal/analyzer/stale.go / waste.go — StaleResult / WasteDir にも PhysicalSize を追加
- [x] internal/analyzer/analyze.go — Analyze 内で各集計を新フィールドで埋める。SavingsEstimate 相当は物理サイズベースで計算する

### 9-C: 表示への反映

- [x] internal/formatter/table.go — 論理サイズと物理サイズの乖離が 10% 以上の場合のみ、サイズ列に `926.4 GB (実 43.2 GB)` 形式で括弧表示 + テスト。乖離が小さい行は従来どおり単一値
- [x] internal/formatter/summary.go — 推定節約可能量を物理サイズで計算するように変更。論理サイズとの差を「スパース節約のまやかし」として注記
- [x] internal/formatter/templates/report.html — treemap / 巨大ファイル Top で「論理 / 物理」切替トグルを追加。デフォルトは論理（見た目の支配度を保つ）、物理に切り替えると実占有ベースの矩形サイズ / 並びになる
- [x] internal/formatter/json.go — Report / buildReport に physical_size 系フィールドを追加（既存 size は維持、差分モード互換確保）

### 9-D: CLI フラグ

- [x] internal/app/app.go — Config に SizeMode（"logical" / "physical"）を追加。`-size-mode` フラグで切替
- [x] デフォルトは "logical"。ただしサマリの「推定節約可能量」だけは常に物理サイズで計算（嘘にならないように）
- [x] README.md にスパース対応の説明セクション追加。Docker.raw / diffdisk の実例を掲載

### 9-E: 検証

- [x] `./mogura ~` を実行し、Docker.raw の物理サイズが `du -sh` と一致することを確認
- [x] `./mogura -diff` が既存の JSON スナップショット（physical_size フィールド無し）を読み込んでも壊れないことを確認（後方互換）


## Phase 10: 高速化（並列 walk + OS 最適化 syscall）

> リサーチ結果: 同種ツール（gdu, dust, DaisyDisk 等）が速く見える正体は「キャッシュ」ではなく「並列 walk + OS ネイティブな一括 stat 取得」だった。mtime 増分キャッシュは POSIX の mtime 伝播ルール（孫以下の変更は親に伝播しない）により素朴実装ではサイレントに古い結果を返すため採用しない。まず D 案（並列 + getattrlistbulk）で初回スキャンを 5〜10 倍速くする。

### 10-A: 並列 walk（依存ゼロ）

> `filepath.WalkDir` は単一 goroutine でディレクトリを順次走査するためディスク I/O 待ちがボトルネック。worker pool + ディレクトリタスクキューに書き換え、SSD で 3〜5 倍の高速化を狙う。scanner 内部の差し替えだけで済み、外部依存は増えない。

- [x] internal/scanner/parallel.go — 新ファイルを作成。ディレクトリタスク用 channel と sync.WaitGroup を用いた worker pool の骨格を実装（worker 数は runtime.NumCPU() デフォルト、ScanOpts.Workers でオーバーライド可）
- [x] internal/scanner/scanner.go — ScanOpts に Workers int フィールドを追加（0 なら NumCPU()）
- [x] internal/scanner/parallel.go — ディレクトリ 1 つを処理する workerFn: os.ReadDir で子を列挙 → ファイルは FileInfo に変換して結果 channel に送る / サブディレクトリはタスク channel に投入 / 除外・symlink スキップ・OneFileSystem・permission warning は既存 scanner.go と同じセマンティクスを保つ
- [ ] internal/scanner/parallel.go — 結果集約: 別 goroutine で結果 channel から FileInfo を受け取り []FileInfo にまとめる。OnProgress コールバックもこの集約側から叩く（現状と同じ挙動を維持）
- [ ] internal/scanner/parallel.go — タスクキューが空になり全 worker が idle になった時点で終了する仕組み（WaitGroup + close(taskCh) の順序に注意。deadlock しないこと）
- [ ] internal/scanner/scanner.go — Scan 関数を並列実装に差し替え。既存の逐次実装は削除（Git 履歴に残る）
- [ ] internal/scanner/parallel_test.go — 並列版でも既存 scanner_test.go のテスト（exclude、symlink、permission、OneFileSystem、PhysicalSize 等）が全てパスすること。map 順非依存の検証を追加
- [ ] internal/scanner/parallel_test.go — 並列性テスト: 数千ファイルを含む一時ツリーを作り、Workers=1 と Workers=8 で同一結果になることを確認
- [ ] bench_test.go（プロジェクトルート or scanner 配下）— 大規模ディレクトリ用のベンチマーク。`go test -bench` で逐次版 vs 並列版のスループットを測れるようにする。CI では走らせず手動実行でよい
- [ ] 実機ベンチ: `./mogura ~` を 3 回実行し、逐次版との差を README に記録（ミニマムな数値報告でよい）
- [ ] README.md に並列スキャンの説明と `-workers` フラグ（internal/app/app.go 側も併せて追加）を追記

### 10-B: getattrlistbulk（darwin 専用の一括 stat）

> `getattrlistbulk(2)` は 1 syscall で 1 ディレクトリ内の全エントリの stat を返す macOS 10.10+ の API。readdir + lstat ループに比べ syscall 数が 1/N になり、APFS で 3〜8 倍の高速化が期待できる。Linux は将来 10-C として statx + getdents64 で同等のことをやる（本 Phase ではやらない）。**外部依存ポリシー要判断**: `golang.org/x/sys/unix` を導入するか、`syscall.Syscall6` で直接叩くか、Phase 10-B 着手時点で決める。

- [ ] **方針決定タスク**: `golang.org/x/sys` を外部依存として許容するか、`syscall.Syscall6` + `unsafe.Pointer` で直接 ABI を叩くかを決める。CLAUDE.md の「外部依存ゼロ」方針の扱いも同時に再確認（OS 特化の syscall ラッパは事実上の標準扱いにするのが現実的。判断を CLAUDE.md に追記）
- [ ] internal/scanner/bulkstat_darwin.go — darwin build tag (`//go:build darwin`) で getattrlistbulk ラッパを実装。`struct attrlist` / `ATTR_BIT_MAP_COUNT` / `ATTR_CMN_NAME` / `ATTR_CMN_OBJTYPE` / `ATTR_CMN_MODTIME` / `ATTR_FILE_TOTALSIZE` / `ATTR_FILE_ALLOCSIZE`（= 物理サイズ、Phase 9 と連携）を取得
- [ ] internal/scanner/bulkstat_darwin.go — readDirBulk(path string) ([]bulkEntry, error) 関数として公開。1 回の呼び出しで 1 バッファ分（通常 64〜256 エントリ）、eof まで繰り返し呼ぶ。属性バッファのパース（可変長レコードのため offset 計算に注意）をユニットテストでカバー
- [ ] internal/scanner/bulkstat_other.go — darwin 以外の build tag (`//go:build !darwin`) で readDirBulk をスタブ実装（`os.ReadDir` + `os.Lstat` にフォールバック）。型とシグネチャを darwin 版と揃える
- [ ] internal/scanner/parallel.go — workerFn を readDirBulk ベースに書き換え。darwin では 1 syscall で済み、他 OS では従来どおり os.ReadDir + Lstat
- [ ] internal/scanner/bulkstat_darwin_test.go — darwin 限定テスト（build tag で制限）。一時ディレクトリに各種ファイル（regular / dir / symlink / スパース）を作り、readDirBulk の結果が Lstat と一致することを検証
- [ ] 実機ベンチ: Phase 10-A 単独版 vs 10-A + 10-B darwin 版で `~/Library` のスキャン時間を比較。README に結果記録
- [ ] フォールバック検証: darwin でも `-bulkstat=false` フラグ（あるいは内部切替）で従来経路に戻せることを確認。何か問題が出た時の脱出口として残す
- [ ] CLAUDE.md 更新: 「外部依存ゼロ」の例外として `golang.org/x/sys` を明示（方針決定タスクの結果次第）


## Phase 11: 開発エコシステム（GitHub Actions + Dependabot）

> ローカルの `make quality` だけでなく、push / PR 時に自動で検証が走る CI パイプラインを整備する。さらに dependabot で依存の追従を自動化する（現状は標準ライブラリのみだが、Phase 10-B で `golang.org/x/sys` を入れる可能性があり、その時点で効いてくる）。GitHub Actions 公式アクションはすべて無料枠で動く。

### 11-A: CI ワークフロー（test + quality）

- [ ] `.github/workflows/ci.yml` 作成。trigger は `push`（main）と `pull_request`。Go 版数は go.mod の `go` ディレクティブから自動取得（`actions/setup-go` の `go-version-file: go.mod`）
- [ ] ジョブ内容: `go vet ./...` → `gofmt -l ./...`（差分あれば fail）→ `go test -race ./...` → `go build ./...`。これは `make quality` とほぼ同じだが `-race` を追加して競合検出も同時に回す
- [ ] matrix で `ubuntu-latest` と `macos-latest` の両方で回す（Phase 10-B の darwin build tag 検証のため macOS 必須）
- [ ] キャッシュ設定: `actions/setup-go` の組み込みキャッシュ（`cache: true`）で go modules と build cache を保持。CI 時間短縮

### 11-B: golangci-lint 導入（任意・品質ゲート強化）

- [ ] `.golangci.yml` を作成。有効化する linter は errcheck / govet / ineffassign / staticcheck / unused / gosimple / misspell あたりから開始。厳しすぎると PR が通らないので段階導入
- [ ] `.github/workflows/ci.yml` に golangci-lint ジョブを追加（`golangci/golangci-lint-action@v6` 公式アクション使用）
- [ ] ローカルでも `make lint` ターゲットを Makefile に追加（CI と同じ設定で走るように）

### 11-C: リリースワークフロー（タグ駆動）

- [ ] `.github/workflows/release.yml` 作成。trigger は `push.tags: 'v*'`
- [ ] `goreleaser-action` を使って darwin/amd64, darwin/arm64, linux/amd64, linux/arm64 のバイナリをビルド・Release に添付
- [ ] `.goreleaser.yml` を作成。CLAUDE.md の「外部依存ゼロ」方針は build 時の話なので goreleaser 自体は問題なし
- [ ] （将来）Homebrew Tap 対応: goreleaser の `brews` セクションで自動 formula 生成（別リポジトリ `homebrew-mogura` が必要、今は延期）

### 11-D: Dependabot

- [ ] `.github/dependabot.yml` 作成
- [ ] `gomod` エコシステムを有効化（`directory: "/"`, `schedule.interval: weekly`）。現状は依存ゼロだが将来 `golang.org/x/sys` を入れる可能性があるので先に設定しておく
- [ ] `github-actions` エコシステムを有効化（`actions/setup-go` 等のバージョン追従）。これは今日から効く
- [ ] PR の自動マージまでは設定しない（人間レビュー必須）。ラベル `dependencies` だけ付与する設定

### 11-E: バッジと README 整備

- [ ] README.md の冒頭に CI バッジ（`https://github.com/<user>/mogura/actions/workflows/ci.yml/badge.svg`）を追加
- [ ] CONTRIBUTING.md（新規）に「PR 前に `make quality` を通すこと」「CI は ubuntu + macOS で回る」と明記
- [ ] `make quality` と CI の内容が一致していることをドキュメントで保証


---

## Backlog

- TestBuildTreeDominantCategory がフラッキーだった（同サイズカテゴリの map イテレーション順依存）。テストデータは修正済みだが、tree.go の DominantCategory 決定ロジック自体に同率時のタイブレークルールを追加すべきか検討

---

## 設計メモ