# Release model

Liki uses **one date-based version** for runtime, contracts, skill metadata, and
product releases. There is no parallel SemVer version.

The operational release pipeline is documented in
[Release pipeline](./RELEASE_PIPELINE.md). Ordinary commits do not build images;
release workflows build, test, sign, and publish immutable image digests.

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
2. Build and publish a candidate image set from the approved source commit.
3. Run smoke, assembly, security, and e2e gates against the candidate digests.
4. If every candidate gate passes, promote the same digests to release tags;
   do not rebuild them.
5. Run `make version` so every date-version surface is updated together.
6. Confirm `CHANGELOG.md` contains the milestone changes.
7. Create an annotated tag on the approved commit:

   ```bash
   git tag -a v2026.10.12.0 <commit> -m "Liki 2026.10.12.0"
   git push origin v2026.10.12.0
   ```

8. Publish the GitHub Release with that tag.
9. Verify release CI, images, signatures, signed release manifest, and artifacts.
