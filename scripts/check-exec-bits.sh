#!/usr/bin/env bash
# Verify git executable bits for scripts invoked without an interpreter prefix.
#
# `bash script.sh` and bare `scripts/x.py` calls (e.g. "$ROOT/scripts/x.py") fail
# with "Permission denied" (make exit 126) once committed with mode 100644. The
# working tree can still be executable while the index is not, so this checks the
# index mode (what git will publish), not `[ -x ]`. Ruff does not check file
# modes, so this covers that blind spot for `make check`.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

fail=0

mode_of() {
  git ls-files -s -- "$1" | awk 'NR==1 {print $1}'
}

require_755() {
  local path="$1" mode
  mode="$(mode_of "$path")"
  if [ -z "$mode" ]; then
    echo "❌ referenced but not tracked: $path" >&2
    fail=1
    return
  fi
  if [ "$mode" != "100755" ]; then
    echo "❌ tracked mode $mode (want 100755): $path" >&2
    fail=1
  fi
}

# Every tracked shell script must be executable.
while IFS= read -r path; do
  [ -n "$path" ] || continue
  require_755 "$path"
done < <(git ls-files -- 'scripts/*.sh')

# Python scripts referenced bare (previous token is not an interpreter) must be
# executable; `python3 scripts/x.py` references are exempt.
while IFS= read -r path; do
  [ -n "$path" ] || continue
  require_755 "$path"
done < <(
  awk '
    {
      n = split($0, tok, /[[:space:]=]+"?/)
      for (i = 1; i <= n; i++) {
        t = tok[i]
        if (t !~ /scripts\/[A-Za-z_0-9]+\.py"?$/) continue
        gsub(/^"/, "", t)
        gsub(/"$/, "", t)
        prev = (i > 1) ? tok[i - 1] : ""
        if (prev ~ /(python[0-9.]*|PY|[Pp]ython)$/) continue
        print t
      }
    }
  ' scripts/*.sh Makefile | sort -u
)

if [ "$fail" -ne 0 ]; then
  echo "exec-bit check failed" >&2
  exit 1
fi
echo "✓ script executable bits"
