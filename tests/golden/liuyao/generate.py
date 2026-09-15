#!/usr/bin/env python3
"""Generate the full Jing Fang eight-palace 64-hexagram core fixture."""
import json
from pathlib import Path
ROOT=Path(__file__).resolve().parents[3]
source=json.loads((ROOT/"engine/internal/engine/liuyao/data/hexagrams.json").read_text())
palaces=source["palaces"]
if len(palaces)!=8 or len(source["hexagrams"])!=64:
    raise SystemExit("source table is not 8x64")
expected=[{"index":i,"name":h["name"],"palace":h["palace"],"shi_pos":h["shi_pos"]} for i,h in enumerate(source["hexagrams"])]
doc={
 "version":1,
 "basis":"京房八宫卦序；每宫本宫、一世至五世、游魂、归魂。",
 "policy":{"palace_sequence":palaces,"world_response":"world_plus_two_positions"},
 "scope":{"palaces":8,"hexagrams_per_palace":8,"case_count":len(expected)},
 "cases":expected,
}
out=ROOT/"engine/internal/engine/liuyao/testdata/core_64_golden.json"
out.write_text(json.dumps(doc,ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
print(f"wrote {len(expected)} cases")
