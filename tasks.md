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

<!-- Phase 1 完了後に展開 -->

## Phase 3: 出力フォーマット

<!-- Phase 2 完了後に展開 -->

## Phase 4: ゴミ発見

<!-- Phase 3 完了後に展開 -->

---

## Backlog

---

## 設計メモ