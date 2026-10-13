#!/usr/bin/env bash
# Copy an already system-tested local candidate to GHCR without rebuilding.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MANIFEST="${MANIFEST:-$ROOT/build/candidate/candidate-digests.json}"
TARGET_REGISTRY="${TARGET_REGISTRY:-ghcr.io}"

[ -f "$MANIFEST" ] || { echo "❌ candidate manifest not found: $MANIFEST" >&2; exit 1; }

readarray -t manifest < <(python3 - "$MANIFEST" <<'PY'
import json, sys
data = json.load(open(sys.argv[1], encoding="utf-8"))
required = {"registry", "repository", "candidate_tag", "images"}
if missing := required - data.keys():
    raise SystemExit(f"missing manifest fields: {', '.join(sorted(missing))}")
if data["registry"] == "ghcr.io":
    raise SystemExit("candidate is already stored in ghcr.io")
for name, digest in data["images"].items():
    print(data["registry"], data["repository"], name, digest, data["candidate_tag"])
PY
)
[ ${#manifest[@]} -gt 0 ] || { echo "❌ candidate manifest has no images" >&2; exit 1; }

for item in "${manifest[@]}"; do
  read -r source_registry repository name digest candidate_tag <<<"$item"
  source="$source_registry/$repository/$name@$digest"
  target="$TARGET_REGISTRY/$repository/$name:$candidate_tag"
  echo "── publishing $target ──"
  docker buildx imagetools create -t "$target" "$source"
  raw="$(docker buildx imagetools inspect --raw "$target")"
  actual="sha256:$(printf '%s' "$raw" | sha256sum | cut -d' ' -f1)"
  [ "$actual" = "$digest" ] || {
    echo "❌ published digest mismatch for $target: $actual != $digest" >&2
    exit 1
  }
done

echo "✅ system-tested candidate published to $TARGET_REGISTRY"
