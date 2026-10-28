#!/usr/bin/env python3
"""Generate the notice shipped with the engine's embedded city dataset."""

from __future__ import annotations

import argparse
import hashlib
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
MANIFEST = ROOT / "data" / "sources" / "manifest.json"
OUTPUT = ROOT / "engine" / "internal" / "engine" / "city" / "ATTRIBUTION.txt"

LICENSE_URLS = (
    ("Open Data Commons Open Database License", "https://opendatacommons.org/licenses/odbl/1-0/"),
    ("CC BY 3.0 IGO", "https://creativecommons.org/licenses/by/3.0/igo/"),
    ("CC BY 4.0", "https://creativecommons.org/licenses/by/4.0/"),
    ("CC BY-SA 2.0", "https://creativecommons.org/licenses/by-sa/2.0/"),
    ("CC0 1.0", "https://creativecommons.org/publicdomain/zero/1.0/"),
    ("CC BY 3.0", "https://creativecommons.org/licenses/by/3.0/"),
    ("CC BY 2.5 India", "https://creativecommons.org/licenses/by/2.5/in/"),
    ("CC BY 2.5", "https://creativecommons.org/licenses/by/2.5/"),
    ("Open Government Canada", "https://open.canada.ca/en/open-government-licence-canada"),
    ("swisstopo", "https://www.swisstopo.admin.ch/en/home/meta/conditions/geodata/geo-licence.html"),
    ("Data license Germany", "https://www.govdata.de/dl-de/by-2-0"),
    ("Etalab Open License", "https://www.etalab.gouv.fr/wp-content/uploads/2017/04/ETALAB-Licence-Ouverte-v2.0.pdf"),
    ("Open Government Licence v3", "https://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"),
    ("Singapore Open Data License", "https://data.gov.sg/open-data-licence"),
)


def license_url(license_name: str) -> str:
    for needle, url in LICENSE_URLS:
        if needle.lower() in license_name.lower():
            return url
    return ""


def attribution(source: dict[str, object]) -> str:
    value = str(source.get("attribution") or "").strip()
    if value:
        return value
    name = str(source.get("name") or "")
    if name.startswith("geoboundaries-"):
        return "GeoBoundaries gbOpen (https://www.geoboundaries.org/)"
    return ""


def generate() -> str:
    manifest_bytes = MANIFEST.read_bytes()
    manifest = json.loads(manifest_bytes)
    manifest_hash = hashlib.sha256(manifest_bytes).hexdigest()
    lines = [
        "Liki engine bundled geographic data notices",
        "================================================",
        "",
        "The repository MIT license applies to Liki code. It does not relicense",
        "the bundled geographic dataset or its third-party sources.",
        "",
        "The generated dataset is built from the sources listed below. Each source",
        "remains subject to its own license and attribution terms. Complete file",
        "hashes and retrieval timestamps are in data/sources/manifest.json.",
        "",
        f"Source manifest SHA-256: {manifest_hash}",
        f"Source manifest generated at: {manifest.get('generated_at', 'unknown')}",
        "",
        "Source inventory",
        "----------------",
        "",
    ]

    for source in manifest.get("sources", []):
        name = str(source.get("name") or "(unnamed)")
        license_name = str(source.get("license") or "(license not recorded)")
        url = str(source.get("url") or "")
        notice = attribution(source)
        link = license_url(license_name)

        lines.extend((f"### {name}", f"License: {license_name}"))
        if link:
            lines.append(f"License URL: {link}")
        if notice:
            lines.append(f"Attribution: {notice}")
        if url:
            lines.append(f"Source: {url}")
        lines.append("")

    return "\n".join(lines).rstrip() + "\n"


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="verify that the generated notice is current")
    args = parser.parse_args()

    generated = generate()
    if args.check:
        current = OUTPUT.read_text(encoding="utf-8") if OUTPUT.is_file() else ""
        if current != generated:
            print(
                "ATTRIBUTION.txt is stale; run scripts/generate_city_attribution.py",
                file=sys.stderr,
            )
            return 1
        return 0

    OUTPUT.write_text(generated, encoding="utf-8")
    print(f"wrote {OUTPUT.relative_to(ROOT)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
