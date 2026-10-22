#!/usr/bin/env python3
"""Build the runtime city index from raw sources.

Raw sources (read as-is, never edited):
  data/sources/datav/{adcode}_full.json          China admin boundaries (DataV)
  data/sources/geonames/cities15000.zip          world cities (GeoNames)
  data/sources/geoboundaries/{ISO3}/...          world admin bboxes (GeoBoundaries)

Output: engine/internal/engine/city/cities_data.json (generated, do not edit).
"""
import argparse
import json
import math
import time
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SRC = ROOT / "data" / "sources"
OUT = ROOT / "engine" / "internal" / "engine" / "city" / "cities_data.json"
TZ_CN = "Asia/Shanghai"

ISO2 = {
    "中国台湾": "TW", "中国澳门": "MO", "中国香港": "HK", "以色列": "IL",
    "俄罗斯": "RU", "加拿大": "CA", "南非": "ZA", "印度": "IN",
    "印度尼西亚": "ID", "土耳其": "TR", "埃及": "EG", "墨西哥": "MX",
    "奥地利": "AT", "孟加拉国": "BD", "尼日利亚": "NG", "尼泊尔": "NP",
    "巴西": "BR", "德国": "DE", "意大利": "IT", "摩洛哥": "MA",
    "斯里兰卡": "LK", "新加坡": "SG", "新西兰": "NZ", "日本": "JP",
    "智利": "CL", "柬埔寨": "KH", "比利时": "BE", "沙特阿拉伯": "SA",
    "法国": "FR", "泰国": "TH", "澳大利亚": "AU", "瑞士": "CH",
    "缅甸": "MM", "美国": "US", "老挝": "LA", "肯尼亚": "KE",
    "英国": "GB", "荷兰": "NL", "菲律宾": "PH", "葡萄牙": "PT",
    "西班牙": "ES", "越南": "VN", "阿根廷": "AR", "阿联酋": "AE",
    "韩国": "KR", "马来西亚": "MY",
}

ISO3 = {
    "中国台湾": "TWN", "中国澳门": "MAC", "中国香港": "HKG", "以色列": "ISR",
    "俄罗斯": "RUS", "加拿大": "CAN", "南非": "ZAF", "印度": "IND",
    "印度尼西亚": "IDN", "土耳其": "TUR", "埃及": "EGY", "墨西哥": "MEX",
    "奥地利": "AUT", "孟加拉国": "BGD", "尼日利亚": "NGA", "尼泊尔": "NPL",
    "巴西": "BRA", "德国": "DEU", "意大利": "ITA", "摩洛哥": "MAR",
    "斯里兰卡": "LKA", "新加坡": "SGP", "新西兰": "NZL", "日本": "JPN",
    "智利": "CHL", "柬埔寨": "KHM", "比利时": "BEL", "沙特阿拉伯": "SAU",
    "法国": "FRA", "泰国": "THA", "澳大利亚": "AUS", "瑞士": "CHE",
    "缅甸": "MMR", "美国": "USA", "老挝": "LAO", "肯尼亚": "KEN",
    "英国": "GBR", "荷兰": "NLD", "菲律宾": "PHL", "葡萄牙": "PRT",
    "西班牙": "ESP", "越南": "VNM", "阿根廷": "ARG", "阿联酋": "ARE",
    "韩国": "KOR", "马来西亚": "MYS",
}

SKIP_PROVINCES = {"710000", "810000", "820000"}


def load_json(path):
    return json.loads(Path(path).read_text(encoding="utf-8"))


def props_center(props):
    c = props.get("center") or props.get("centroid")
    if isinstance(c, (list, tuple)) and len(c) == 2:
        return float(c[0]), float(c[1])
    return None


def geom_bbox(geom):
    if not geom:
        return None
    kind = geom.get("type")
    if kind == "GeometryCollection":
        boxes = [geom_bbox(g) for g in geom.get("geometries", [])]
        boxes = [b for b in boxes if b]
        if not boxes:
            return None
        return (min(b[0] for b in boxes), max(b[1] for b in boxes),
                min(b[2] for b in boxes), max(b[3] for b in boxes))
    coords = geom.get("coordinates")
    if not coords:
        return None
    west, east, south, north = 180.0, -180.0, 90.0, -90.0
    found = False

    def walk(node):
        nonlocal west, east, south, north, found
        if isinstance(node, (list, tuple)):
            if len(node) >= 2 and isinstance(node[0], (int, float)) and isinstance(node[1], (int, float)):
                x, y = float(node[0]), float(node[1])
                west, east = min(west, x), max(east, x)
                south, north = min(south, y), max(north, y)
                found = True
            else:
                for item in node:
                    walk(item)
    walk(coords)
    if not found:
        return None
    return (west, east, south, north)


def ring_contains(x, y, ring):
    inside = False
    n = len(ring)
    j = n - 1
    for i in range(n):
        xi, yi = ring[i][0], ring[i][1]
        xj, yj = ring[j][0], ring[j][1]
        if ((yi > y) != (yj > y)) and (x < (xj - xi) * (y - yi) / ((yj - yi) or 1e-12) + xi):
            inside = not inside
        j = i
    return inside


def geom_contains(geom, x, y):
    if not geom:
        return False
    kind = geom.get("type")
    if kind == "Polygon":
        rings = geom.get("coordinates") or []
        return bool(rings) and ring_contains(x, y, rings[0])
    if kind == "MultiPolygon":
        for poly in geom.get("coordinates") or []:
            if poly and ring_contains(x, y, poly[0]):
                return True
    if kind == "GeometryCollection":
        return any(geom_contains(g, x, y) for g in geom.get("geometries", []))
    return False


def bbox_dict(b):
    if not b:
        return None
    west, east, south, north = b
    if east < west:  # antimeridian; omit rather than mislead
        return None
    return {"west": round(west, 6), "east": round(east, 6),
            "south": round(south, 6), "north": round(north, 6)}


def build_china():
    counties, cities = [], []
    top = load_json(SRC / "datav" / "100000_full.json")
    for prov in top.get("features", []):
        p = prov.get("properties") or {}
        pname = (p.get("name") or "").strip()
        pad = str(p.get("adcode") or "")
        if not pname or not pad or pad in ("100000",) or pad in SKIP_PROVINCES:
            continue
        pfile = SRC / "datav" / f"{pad}_full.json"
        if not pfile.exists():
            print(f"  skip province {pname} ({pad}): no raw file", flush=True)
            continue
        pdata = load_json(pfile)
        feats = pdata.get("features", [])
        if not feats:
            continue
        levels = {(f.get("properties") or {}).get("level") for f in feats}
        if levels <= {"district"}:
            # municipality: province itself is the city; features are its districts
            c = props_center(p)
            city = {"name": pname, "region": pname}
            if c:
                city["lng"], city["lat"] = c
            b = bbox_dict(geom_bbox(prov.get("geometry")))
            if b:
                city["bbox"] = b
            city["tz"] = TZ_CN
            cities.append(city)
            for f in feats:
                q = f.get("properties") or {}
                name = (q.get("name") or "").strip()
                if not name:
                    continue
                rec = {"name": name, "region": pname}
                qc = props_center(q)
                if qc:
                    rec["lng"], rec["lat"] = qc
                qb = bbox_dict(geom_bbox(f.get("geometry")))
                if qb:
                    rec["bbox"] = qb
                rec["tz"] = TZ_CN
                counties.append(rec)
            continue
        for f in feats:
            q = f.get("properties") or {}
            name = (q.get("name") or "").strip()
            qad = str(q.get("adcode") or "")
            if not name or not qad:
                continue
            if "直辖县级行政区划" in name:
                sub = SRC / "datav" / f"{qad}_full.json"
                if sub.exists():
                    for d in load_json(sub).get("features", []):
                        dq = d.get("properties") or {}
                        dname = (dq.get("name") or "").strip()
                        if not dname:
                            continue
                        rec = {"name": dname, "region": pname}
                        dc = props_center(dq)
                        if dc:
                            rec["lng"], rec["lat"] = dc
                        db = bbox_dict(geom_bbox(d.get("geometry")))
                        if db:
                            rec["bbox"] = db
                        rec["tz"] = TZ_CN
                        counties.append(rec)
                continue
            center = props_center(q)
            bbox = bbox_dict(geom_bbox(f.get("geometry")))
            if qad.endswith("00"):
                city = {"name": name, "region": pname}
                if center:
                    city["lng"], city["lat"] = center
                if bbox:
                    city["bbox"] = bbox
                city["tz"] = TZ_CN
                cities.append(city)
                children = int(q.get("childrenNum") or 0)
                if children > 0:
                    sub = SRC / "datav" / f"{qad}_full.json"
                    if sub.exists():
                        for d in load_json(sub).get("features", []):
                            dq = d.get("properties") or {}
                            dname = (dq.get("name") or "").strip()
                            if not dname:
                                continue
                            rec = {"name": dname, "region": pname, "city": name}
                            dc = props_center(dq)
                            if dc:
                                rec["lng"], rec["lat"] = dc
                            db = bbox_dict(geom_bbox(d.get("geometry")))
                            if db:
                                rec["bbox"] = db
                            rec["tz"] = TZ_CN
                            counties.append(rec)
                else:
                    rec = {"name": name, "region": pname}
                    if center:
                        rec["lng"], rec["lat"] = center
                    if bbox:
                        rec["bbox"] = bbox
                    rec["tz"] = TZ_CN
                    counties.append(rec)
            else:
                # province-direct county-level city (e.g. 济源市) or district
                rec = {"name": name, "region": pname}
                if center:
                    rec["lng"], rec["lat"] = center
                if bbox:
                    rec["bbox"] = bbox
                rec["tz"] = TZ_CN
                counties.append(rec)
    return counties, cities


def load_geonames():
    z = zipfile.ZipFile(SRC / "geonames" / "cities15000.zip")
    by_country = {}
    for raw in z.open("cities15000.txt"):
        f = raw.decode("utf-8").rstrip("\n").split("\t")
        if len(f) < 18:
            continue
        cc = f[8]
        row = {
            "name": f[1], "asciiname": f[2], "alts": f[3],
            "lat": float(f[4]), "lon": float(f[5]),
            "population": int(f[14] or 0), "tz": f[17],
        }
        by_country.setdefault(cc, []).append(row)
    return by_country


def haversine(lat1, lon1, lat2, lon2):
    r = 6371.0
    p1, p2 = math.radians(lat1), math.radians(lat2)
    dp = math.radians(lat2 - lat1)
    dl = math.radians(lon2 - lon1)
    a = math.sin(dp / 2) ** 2 + math.cos(p1) * math.cos(p2) * math.sin(dl / 2) ** 2
    return 2 * r * math.asin(math.sqrt(a))


def nearest_row(rows, lat, lon):
    best, bestd = None, 1e18
    for row in rows:
        d = haversine(lat, lon, row["lat"], row["lon"])
        if d < bestd:
            best, bestd = row, d
    return best, bestd


def is_cjk(s):
    return any("\u3400" <= ch <= "\u9fff" for ch in s)


def aliases_from(row):
    """Canonical/English/Chinese aliases; skip the multilingual alphabetical noise."""
    out, seen = [], set()

    def add(s):
        s = s.strip()
        if not s or len(s) > 40 or s in seen or not any(ch.isalpha() for ch in s):
            return
        seen.add(s)
        out.append(s)

    add(row["name"])
    add(row["asciiname"])
    for s in row["alts"].split(","):
        s = s.strip()
        if not s:
            continue
        if is_cjk(s):
            add(s)
        elif s.isascii() and (row["name"].startswith(s) or s.startswith(row["name"])
                              or (len(s) <= 4 and s.isalpha())):
            add(s)
        if len(out) >= 16:
            break
    return out


def name_matched_row(rows, name, lat, lon):
    """Prefer a GeoNames row whose names contain the world's Chinese name."""
    if not name or len(name) < 2:
        return None
    best, bestd = None, 1e18
    for row in rows:
        names = [row["name"], row["asciiname"]] + row["alts"].split(",")
        if not any(n == name or (n.startswith(name) and len(n) <= len(name) + 3) for n in names):
            continue
        d = haversine(lat, lon, row["lat"], row["lon"])
        if d < bestd:
            best, bestd = row, d
    return best if best is not None and bestd <= 200 else None


_gb_cache = {}
_datav_top = None


def gb_features(iso, adm):
    key = (iso, adm)
    if key in _gb_cache:
        return _gb_cache[key]
    path = SRC / "geoboundaries" / iso / f"{iso}-{adm}-simplified.geojson"
    feats = load_json(path).get("features", []) if path.exists() else []
    _gb_cache[key] = feats
    return feats


def gb_bbox(iso, lon, lat):
    for adm in ("ADM2", "ADM1"):
        best = None
        best_area = None
        for f in gb_features(iso, adm):
            if not geom_contains(f.get("geometry"), lon, lat):
                continue
            b = geom_bbox(f.get("geometry"))
            if not b:
                continue
            area = (b[1] - b[0]) * (b[3] - b[2])
            if best is None or area < best_area:
                best, best_area = b, area
        if best is not None:
            return bbox_dict(best)
    return None


def datav_territory_bbox(region):
    """DataV top-level geometry for HK/Macau (GeoBoundaries has no rows there)."""
    global _datav_top
    names = {"中国香港": "香港特别行政区", "中国澳门": "澳门特别行政区"}
    want = names.get(region)
    if not want:
        return None
    if _datav_top is None:
        path = SRC / "datav" / "100000_full.json"
        _datav_top = load_json(path).get("features", []) if path.exists() else []
    for f in _datav_top:
        p = f.get("properties") or {}
        if (p.get("name") or "").strip() == want:
            return bbox_dict(geom_bbox(f.get("geometry")))
    return None


def build_world(current):
    geo = load_geonames()
    out = []
    warnings = []
    for rec in current:
        region = rec.get("region", "")
        cc = ISO2.get(region, "")
        iso3 = ISO3.get(region, "")
        rows = geo.get(cc, [])
        row, dist = nearest_row(rows, rec["lat"], rec["lng"]) if rows else (None, 1e18)
        named = name_matched_row(rows, rec["name"], rec["lat"], rec["lng"])
        alias_row = named or (row if row and dist <= 120 else None)
        updated = dict(rec)
        if row and dist <= 120:
            updated["tz"] = row["tz"]
        elif not updated.get("tz"):
            warnings.append(f"{rec['name']} ({region}): no GeoNames match within 120km (nearest {dist:.0f}km)")
        if alias_row:
            aliases = aliases_from(alias_row)
            if aliases:
                updated["aliases"] = aliases
        b = gb_bbox(iso3, rec["lng"], rec["lat"]) if iso3 else None
        if not b:
            b = datav_territory_bbox(region)
        if b:
            updated["bbox"] = b
        out.append(updated)
    return out, warnings


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--dry-run", action="store_true")
    args = ap.parse_args()

    current = load_json(OUT)
    counties, cities = build_china()
    world, warnings = build_world(current.get("world", []))

    print(f"china: counties={len(counties)} (was {len(current.get('counties', []))}), "
          f"cities={len(cities)} (was {len(current.get('cities', []))}), "
          f"world={len(world)}", flush=True)
    for w in warnings:
        print("  WARN", w, flush=True)
    missing_bbox = [r["name"] for r in world if not r.get("bbox")]
    print(f"world without bbox: {len(missing_bbox)} {missing_bbox[:10]}", flush=True)
    missing_tz = [r["name"] for r in world if not r.get("tz")]
    print(f"world without tz: {len(missing_tz)} {missing_tz[:10]}", flush=True)

    data = {
        "counties": counties,
        "cities": cities,
        "world": world,
        "meta": {
            "generated_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            "sources": {
                "china": "DataV.GeoAtlas areas_v3 (data/sources/datav)",
                "world_cities": "GeoNames cities15000 (data/sources/geonames)",
                "world_bbox": "GeoBoundaries gbOpen (data/sources/geoboundaries)",
                "manifest": "data/sources/manifest.json",
            },
            "counts": {"counties": len(counties), "cities": len(cities), "world": len(world)},
        },
    }
    if args.dry_run:
        print("dry-run: not writing", flush=True)
        return
    OUT.write_text(json.dumps(data, ensure_ascii=False, separators=(",", ":")) + "\n", encoding="utf-8")
    print(f"wrote {OUT.relative_to(ROOT)} ({OUT.stat().st_size/1e6:.1f} MB)", flush=True)


if __name__ == "__main__":
    main()
