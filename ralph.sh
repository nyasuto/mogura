#!/usr/bin/env bash
# ralph.sh — タスク駆動型自律開発ループ
# Usage: ./ralph.sh [max_iterations]
set -euo pipefail

MAX=${1:-50}
DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="$DIR/.ralph-logs"
mkdir -p "$LOG_DIR"
RUN_ID="$(date +%Y%m%d-%H%M%S)"

for i in $(seq 1 "$MAX"); do
  remaining=$(grep -c '^- \[ \]' "$DIR/tasks.md" 2>/dev/null || echo 0)
  if [[ "$remaining" -eq 0 ]]; then
    echo "✅ All tasks complete."
    break
  fi
  log="$LOG_DIR/${RUN_ID}-iter${i}.log"
  echo "── iteration $i/$MAX (remaining: $remaining) → $log ──"
  claude --print --dangerously-skip-permissions "$(cat "$DIR/PROMPT.md")" 2>&1 | tee "$log"
  sleep 3
done