#!/usr/bin/env python3
"""Download third-party source data as-is (raw) with provenance.

Sources:
  - GeoNames cities15000.zip (CC BY 4.0)
  - GeoBoundaries gbOpen ADM2/ADM1 simplified GeoJSON (CC BY 4.0 per boundary)
  - DataV China admin GeoJSON (fetched by a separate step; see --china)

Writes data/sources/manifest.json with URL, version, license, attribution,
sha256, bytes and retrieval time for every file.
"""
import argparse
import hashlib
import json
import time
import urllib.error
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SOURCES = ROOT / "data" / "sources"
MANIFEST = SOURCES / "manifest.json"
UA = "liki-source-fetch/1.0"

GEONAMES_URL = "https://download.geonames.org/export/dump/cities15000.zip"
GEONAMES_LICENSE = "CC BY 4.0"
GEONAMES_ATTRIBUTION = "GeoNames (https://www.geonames.org/)"

GB_API = "https://www.geoboundaries.org/api/current/gbOpen/{iso}/{adm}/"
GB_LEVELS = ("ADM2", "ADM1")

# Chinese region label in city data -> ISO3 (GeoBoundaries / ISO 3166-1 alpha-3).
ISO3 = {
    "中国台湾": "TWN", "中国澳门": "MAC", "中国香港": "HKG",
    "以色列": "ISR", "俄罗斯": "RUS", "加拿大": "CAN", "南非": "ZAF",
    "印度": "IND", "印度尼西亚": "IDN", "土耳其": "TUR", "埃及": "EGY",
    "墨西哥": "MEX", "奥地利": "AUT", "孟加拉国": "BGD", "尼日利亚": "NGA",
    "尼泊尔": "NPL", "巴西": "BRA", "德国": "DEU", "意大利": "ITA",
    "摩洛哥": "MAR", "斯里兰卡": "LKA", "新加坡": "SGP", "新西兰": "NZL",
    "日本": "JPN", "智利": "CHL", "柬埔寨": "KHM", "比利时": "BEL",
    "沙特阿拉伯": "SAU", "法国": "FRA", "泰国": "THA", "澳大利亚": "AUS",
    "瑞士": "CHE", "缅甸": "MMR", "美国": "USA", "老挝": "LAO",
    "肯尼亚": "KEN", "英国": "GBR", "荷兰": "NLD", "菲律宾": "PHL",
    "葡萄牙": "PRT", "西班牙": "ESP", "越南": "VNM", "阿根廷": "ARG",
    "阿联酋": "ARE", "韩国": "KOR", "马来西亚": "MYS",
}


def now():
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())


def load_manifest():
    if MANIFEST.exists():
        return json.loads(MANIFEST.read_text(encoding="utf-8"))
    return {"generated_at": now(), "sources": []}


def save_manifest(m):
    m["generated_at"] = now()
    MANIFEST.parent.mkdir(parents=True, exist_ok=True)
    MANIFEST.write_text(json.dumps(m, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def get_json(url, timeout=40):
    req = urllib.request.Request(url, headers={"User-Agent": UA})
    with urllib.request.urlopen(req, timeout=timeout) as r:
        return json.loads(r.read().decode("utf-8"))


def media_url(url):
    """GitHub raw URLs for GeoBoundaries point at LFS files; use the media endpoint."""
    marker = "https://github.com/wmgeolab/geoBoundaries/raw/"
    if url.startswith(marker):
        return "https://media.githubusercontent.com/media/wmgeolab/geoBoundaries/" + url[len(marker):]
    marker2 = "https://raw.githubusercontent.com/wmgeolab/geoBoundaries/"
    if url.startswith(marker2):
        return "https://media.githubusercontent.com/media/wmgeolab/geoBoundaries/" + url[len(marker2):]
    return url


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def download(url, dest, resume=True, retries=3):
    dest.parent.mkdir(parents=True, exist_ok=True)
    part = dest.with_name(dest.name + ".part")
    start = 0
    if resume and part.exists():
        start = part.stat().st_size
    req = urllib.request.Request(url, headers={"User-Agent": UA})
    if start:
        req.add_header("Range", f"bytes={start}-")
    t0 = time.time()
    try:
        with urllib.request.urlopen(req, timeout=60) as r, open(part, "ab" if start else "wb") as f:
            got = 0
            while True:
                chunk = r.read(1 << 16)
                if not chunk:
                    break
                f.write(chunk)
                got += len(chunk)
            dt = max(time.time() - t0, 1e-6)
            part.rename(dest)
            size = dest.stat().st_size
            print(f"  {dest.name}: +{got/1e6:.1f}MB (total {size/1e6:.1f}MB) in {dt:.0f}s "
                  f"({got/dt/1024:.0f} KB/s)", flush=True)
            return size
    except urllib.error.HTTPError as e:
        if e.code == 416 and start and part.exists():
            part.rename(dest)
            print(f"  {dest.name}: resumed as complete ({dest.stat().st_size} bytes)", flush=True)
            return dest.stat().st_size
        if retries > 0:
            print(f"  retry {dest.name} ({e})", flush=True)
            time.sleep(2)
            return download(url, dest, resume=True, retries=retries - 1)
        raise
    except (urllib.error.URLError, TimeoutError, ConnectionError, OSError) as e:
        if retries > 0:
            print(f"  retry {dest.name} ({e})", flush=True)
            time.sleep(2)
            return download(url, dest, resume=True, retries=retries - 1)
        raise


def fetch_geonames(manifest, force=False):
    dest = SOURCES / "geonames" / "cities15000.zip"
    if force or not dest.exists() or dest.stat().st_size < 3_000_000:
        print("GeoNames cities15000", flush=True)
        download(GEONAMES_URL, dest, resume=True)
    entry = {
        "name": "geonames-cities15000",
        "url": GEONAMES_URL,
        "path": str(dest.relative_to(ROOT)),
        "version": "cities15000 (live export)",
        "license": GEONAMES_LICENSE,
        "attribution": GEONAMES_ATTRIBUTION,
        "bytes": dest.stat().st_size,
        "sha256": sha256(dest),
        "retrieved_at": now(),
    }
    manifest["sources"] = [s for s in manifest["sources"] if s.get("name") != entry["name"]] + [entry]
    save_manifest(manifest)


def fetch_geoboundaries(manifest, only=None, force=False):
    isos = [only] if only else sorted(set(ISO3.values()))
    for iso in isos:
        for adm in GB_LEVELS:
            meta_url = GB_API.format(iso=iso, adm=adm)
            try:
                meta = get_json(meta_url)
            except urllib.error.HTTPError as e:
                if e.code == 404:
                    print(f"GeoBoundaries {iso}/{adm}: not available (404)", flush=True)
                    continue
                raise
            meta_dest = SOURCES / "geoboundaries" / iso / f"{iso}-{adm}.meta.json"
            meta_dest.parent.mkdir(parents=True, exist_ok=True)
            meta_dest.write_text(json.dumps(meta, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
            url = media_url(meta.get("simplifiedGeometryGeoJSON") or meta.get("gjDownloadURL"))
            dest = SOURCES / "geoboundaries" / iso / f"{iso}-{adm}-simplified.geojson"
            if force or not dest.exists():
                print(f"GeoBoundaries {iso}/{adm}", flush=True)
                try:
                    download(url, dest, resume=True)
                except Exception as e:
                    print(f"  FAILED {iso}/{adm}: {e}", flush=True)
                    continue
            entry = {
                "name": f"geoboundaries-{iso}-{adm}",
                "url": url,
                "source_url": meta_url,
                "path": str(dest.relative_to(ROOT)),
                "version": meta.get("boundaryYearRepresented", ""),
                "license": meta.get("boundaryLicense", "CC BY 4.0"),
                "license_detail": meta.get("licenseDetail", ""),
                "source": meta.get("boundarySource", ""),
                "build_date": meta.get("buildDate", ""),
                "bytes": dest.stat().st_size,
                "sha256": sha256(dest),
                "retrieved_at": now(),
            }
            manifest["sources"] = [s for s in manifest["sources"] if s.get("name") != entry["name"]] + [entry]
            save_manifest(manifest)


DATAV_ROOT = "https://geo.datav.aliyun.com/areas_v3/bound"
DATAV_LICENSE = "DataV.GeoAtlas (Aliyun) service terms"
DATAV_ATTRIBUTION = "阿里云 DataV.GeoAtlas (https://datav.aliyun.com/portal/school/atlas/area_selector)"


def fetch_datav(manifest, force=False):
    def fetch(adcode):
        url = f"{DATAV_ROOT}/{adcode}_full.json"
        dest = SOURCES / "datav" / f"{adcode}_full.json"
        if force or not dest.exists():
            print(f"DataV {adcode}", flush=True)
            try:
                download(url, dest, resume=True)
            except Exception as e:
                print(f"  FAILED {adcode}: {e}", flush=True)
                return None
        try:
            data = json.loads(dest.read_text(encoding="utf-8"))
        except Exception as e:
            print(f"  BAD JSON {adcode}: {e}", flush=True)
            return None
        entry = {
            "name": f"datav-{adcode}",
            "url": url,
            "path": str(dest.relative_to(ROOT)),
            "version": "areas_v3 (live)",
            "license": DATAV_LICENSE,
            "attribution": DATAV_ATTRIBUTION,
            "bytes": dest.stat().st_size,
            "sha256": sha256(dest),
            "retrieved_at": now(),
        }
        manifest["sources"] = [s for s in manifest["sources"] if s.get("name") != entry["name"]] + [entry]
        save_manifest(manifest)
        return data

    top = fetch("100000")
    if not top:
        return
    for prov in top.get("features", []):
        p = prov.get("properties") or {}
        pname, pad = (p.get("name") or "").strip(), str(p.get("adcode") or "")
        if not pname or not pad or pad == "100000":
            continue
        pdata = fetch(pad)
        if not pdata:
            continue
        for feat in pdata.get("features", []):
            q = feat.get("properties") or {}
            qad, qlevel = str(q.get("adcode") or ""), (q.get("level") or "")
            if not qad or qad == pad:
                continue
            if qlevel == "city" and int(q.get("childrenNum") or 0) > 0:
                fetch(qad)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--foreign", action="store_true", help="GeoNames + GeoBoundaries")
    ap.add_argument("--china", action="store_true", help="DataV China admin GeoJSON")
    ap.add_argument("--only-iso", help="fetch a single ISO3")
    ap.add_argument("--force", action="store_true")
    args = ap.parse_args()
    if not (args.foreign or args.china or args.only_iso):
        ap.error("use --foreign, --china or --only-iso")
    SOURCES.mkdir(parents=True, exist_ok=True)
    m = load_manifest()
    if args.foreign:
        fetch_geonames(m, force=args.force)
        fetch_geoboundaries(m, force=args.force)
    if args.china:
        fetch_datav(m, force=args.force)
    if args.only_iso:
        fetch_geoboundaries(m, only=args.only_iso.upper(), force=args.force)


if __name__ == "__main__":
    main()
