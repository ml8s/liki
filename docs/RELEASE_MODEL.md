# Release model

Liki uses **one date-based version** (`YYYY.MM.DD.SERIAL`) as the source for
runtime, contracts, and product releases. WorkBuddy marketplace artifacts carry a
**derived** semver projection of that version (see below); because it is derived
rather than hand-maintained, there is still no parallel version.

The operational release pipeline is documented in
[Release pipeline](./RELEASE_PIPELINE.md). Neither ordinary commits nor CI
build `liki` images; `release.yml` builds, smokes, and signs them when a
`vYYYY.MM.DD.N` tag is pushed. Publishing the GitHub Release is explicit:
`gh workflow run release-publish.yml -f tag=vYYYY.MM.DD.N` — a tag never
publishes by itself.

## Version format

```text
YYYY.MM.DD.SERIAL
```

Example:

```text
2026.10.12.0
```

A release tag adds the conventional Git `v` prefix:

```text
v2026.10.12.0
```

The tag is a release name only; the version stored in files has no `v` prefix.

## WorkBuddy artifact version (semver projection)

The marketplace parser rejects 4-segment versions (`版本号 "2026.10.16.2" 格式不合法`),
so every submitted artifact carries a projection of the release CalVer:

```text
YYYY.MM.DD.SERIAL  →  YYYY.MMDD.SERIAL      (month/day leading zeros dropped)
2026.10.16.2       →  2026.1016.2
```

The rule is defined once in `scripts/version_projection.py`; `make version` and the
gates both call it, so the artifact version is never hand-written and never drifts.
It stays a legal semver, increases strictly (same day via `SERIAL`, next day via
`MMDD`, next year via `YYYY`), and is reversible (`MINOR` is always 3–4 digits).

Surfaces carrying the projection: `skills/liki/SKILL.md`, `expert-packs/*/SKILL.md`
(identity source), generated `expert-packs/*/skills/*/SKILL.md`, generated
`expert-packs/*/.codebuddy-plugin/plugin.json`, and
`workbuddy/*-connector/connector-meta.json`.

## Single version surface

`make version` updates all runtime version surfaces together:

| Surface | Location |
| --- | --- |
| Runtime skill version | `skills/liki/VERSION.txt` |
| Engine MCP runtime | `engine/cmd/engine-mcp/VERSION` |
| Counsel runtime | `counsel/app/VERSION.txt` |
| Skill package metadata | `skills/liki/SKILL.md` frontmatter `version`（平台必填字段，顶层；写投影值） |
| Expert package metadata | `expert-packs/*/SKILL.md`, nested domain `SKILL.md`, and `.codebuddy-plugin/plugin.json`（写投影值） |
| WorkBuddy connector metadata | `workbuddy/*-connector/connector-meta.json`（写投影值） |
| Counsel projection contracts | `counsel/app/*/tools/*_projection_contract.json` |
| Deployment profiles | `profiles/*.json` |
| Root and counsel package metadata | `pyproject.toml`, `counsel/pyproject.toml` |
| Published engine image reference in README | `README.md`, `README.en.md` |

Rules:

- The serial starts at `0` for each date and increments for additional releases on the same date.
- Version strings remain dotted integers; tests reject other formats.
- Do not introduce or maintain a hand-written semver version. The only semver in
  the repository is the derived WorkBuddy projection above.
- Use the changelog and release notes to describe compatibility, breaking changes, and migration steps; the date version does not encode semantics.

## Cross-repository dependency

Repositories version independently. Liki builds against a locally built
`liki-agent-runtime` image selected by that repository's commit sha; the
assembly target receives it through `BASE_IMAGE`. There is no committed pin
file and no registry pull in the local flow.
