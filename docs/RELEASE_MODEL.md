# Release model

Liki uses **one date-based version** for runtime, contracts, skill metadata, and
product releases. There is no parallel SemVer version.

The operational release pipeline is documented in
[Release pipeline](./RELEASE_PIPELINE.md). Neither ordinary commits nor GitHub
Actions build `liki` images; build, locally validate, and publish candidates with `make
candidate`, then use the promotion workflow to publish, smoke, and sign the
immutable image digests.

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

## Single version surface

`make version` updates all runtime version surfaces together:

| Surface | Location |
| --- | --- |
| Runtime skill version | `skills/liki/VERSION.txt` |
| Engine MCP / RPC runtime | `engine/cmd/engine-mcp/VERSION`, `engine/cmd/engine-rpc/VERSION` |
| Counsel runtime | `counsel/app/VERSION.txt` |
| Skill package metadata | `skills/liki/SKILL.md` frontmatter `metadata.version` |
| Expert package metadata | `expert-packs/*/SKILL.md`, nested domain `SKILL.md`, and `.codebuddy-plugin/plugin.json` |
| Counsel tool / projection contracts | manifests updated by `make version` |
| Deployment profiles | `profiles/*.json` |
| Root and counsel package metadata | `pyproject.toml`, `counsel/pyproject.toml` |
| Published engine image reference in README | `README.md`, `README.en.md` |

Rules:

- The serial starts at `0` for each date and increments for additional releases on the same date.
- Version strings remain dotted integers; tests reject other formats.
- Do not introduce or maintain a parallel SemVer version.
- Use the changelog and release notes to describe compatibility, breaking changes, and migration steps; the date version does not encode semantics.

## Cross-repository dependency pins

Repositories version independently. Liki pins external runtimes explicitly:

| Dependency | Pin file | Locked content |
| --- | --- | --- |
| Agent runtime base | `deps/liki-agents.base` | liki-agents version + immutable digest |

The validation and assembly image must use the same pinned runtime. Do not
bump the pin to an unreleased or dirty build.

## Release process

1. Confirm `main` is green and ordinary commit CI did not build images.
2. Select the next CalVer, run `make version`, update `CHANGELOG.md`, and update
   any released dependency pins. Commit and push this release-preparation change.
3. Build the candidate image set from that clean source commit into the local
   registry and run image smoke.
4. Run assembly, security, and e2e gates against the exact local candidate
   digests.
5. If every candidate gate passes, publish the same digests to GHCR, create the
   release tag on the same source commit, and dispatch promotion for those exact
   digests; do not rebuild them.
6. The promotion workflow publishes the GitHub Release, signs the exact digests,
   and verifies those signatures.
7. Verify release CI, image signatures, signed release manifest, and artifacts.

   For example, create an annotated tag on the approved commit:

   ```bash
   git tag -a v2026.10.12.0 <commit> -m "Liki 2026.10.12.0"
   git push origin v2026.10.12.0
   ```
