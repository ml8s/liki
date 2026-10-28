# Notices

## Code

Liki source code is licensed under the repository [MIT license](./LICENSE).

## Bundled geographic data

The engine embeds a generated geographic dataset in
`engine/internal/engine/city/cities_data.json`. That dataset is **not** relicensed
under the repository MIT license.

It is derived from DataV.GeoAtlas, GeoNames, and GeoBoundaries sources.
`data/sources/manifest.json` is the complete provenance record, including
source URLs, licenses, hashes, and retrieval timestamps.
`engine/internal/engine/city/ATTRIBUTION.txt` is the distributable notice
generated from that manifest.

Engine images carry the generated attribution notice at:

```text
/usr/share/licenses/liki-engine/ATTRIBUTION.txt
```
