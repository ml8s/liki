#!/usr/bin/env bash
# Copy an already system-tested local candidate to GHCR without rebuilding.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MANIFEST="${MANIFEST:-$ROOT/build/candidate/candidate-digests.json}"
TARGET_REGISTRY="${TARGET_REGISTRY:-ghcr.io}"

[ -f "$MANIFEST" ] || { echo "❌ candidate manifest not found: $MANIFEST" >&2; exit 1; }

# Publish only after liki-deploy's exact-digest assembly gate passed.
python3 - "$MANIFEST" <<'PY'
import hashlib
import json
import pathlib
import sys

manifest_path = pathlib.Path(sys.argv[1]).resolve()
evidence_path = manifest_path.with_name("assembly-evidence.json")
if not evidence_path.is_file():
    raise SystemExit(f"❌ assembly evidence not found: {evidence_path}")
try:
    evidence = json.loads(evidence_path.read_text(encoding="utf-8"))
except (OSError, json.JSONDecodeError) as error:
    raise SystemExit(f"❌ invalid assembly evidence {evidence_path}: {error}") from error
manifest_sha256 = hashlib.sha256(manifest_path.read_bytes()).hexdigest()
if evidence.get("status") != "pass":
    raise SystemExit("❌ candidate assembly did not pass")
if evidence.get("gate") != "liki-deploy/candidate-assembly":
    raise SystemExit("❌ assembly evidence is not from the candidate assembly gate")
if evidence.get("manifest_sha256") != manifest_sha256:
    raise SystemExit("❌ assembly evidence does not match this candidate manifest")
print("✓ candidate assembly evidence verified")
PY

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
