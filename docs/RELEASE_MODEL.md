# Release model

Liki has **one hand-maintained release version** in `skills/liki/VERSION.txt`.
Internal package files do not carry release versions. Images are identified by
commit sha; marketplace-only projections are injected when a package is built.

## Version format

```text
YYYY.MM.DD.SERIAL
```

Example: `2026.10.12.0`.

A release tag adds the conventional Git `v` prefix: `v2026.10.12.0`. The tag is
a release name only; the version stored in the file has no `v` prefix.

## Generated version surfaces

`make version` writes only `skills/liki/VERSION.txt`. Other consumers derive
from it:

| Surface | Handling |
| --- | --- |
| Engine MCP runtime | build 注入 |
| Counsel runtime | build 注入 |
| Marketplace skill metadata | packaging injects semver projection |
| WorkBuddy plugin metadata | packaging injects semver projection |
| Agent deployment metadata | generator injects CalVer |
| Internal Python package files | fixed `0.0.0`; not release identity |

Marketplaces reject four-segment versions, so packaging derives:

```text
YYYY.MM.DD.SERIAL  →  YYYY.MMDD.SERIAL
2026.10.16.2       →  2026.1016.2
```

The rule is defined once in `scripts/version_projection.py`. It is never
hand-written in source identity or topology files.

## Cross-repository dependency

Repositories version independently. Liki builds against a locally built
`liki-agent-runtime` image selected by that repository's commit sha; the
assembly target receives it through `BASE_IMAGE`. There is no committed pin
file and no registry pull in the local flow.
