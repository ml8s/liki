# Release model

Liki uses two coordinated version dimensions. They serve different purposes and must not be merged into one version string.

## 1. Runtime / engineering version: CalVer

The executable contract uses a date-stamped version:

```text
YYYY.MM.DD.SERIAL
```

Example:

```text
2026.09.15.7
```

Sources:

- `skills/liki/VERSION`
- `engine/cmd/liki/VERSION`
- `rpc.discover.info.version`
- skill version self-check
- compatibility gates

Rules:

- `make version` writes the same CalVer to the unified skill and engine version files.
- The serial resets to `0` each local date and increments for further releases on that date.
- CalVer is the engineering / compatibility version. It changes for normal development, fixes, docs and feature work.
- Tests and runtime compatibility require this value to remain dotted-integer CalVer. Do not replace it with SemVer.

## 2. Product release: SemVer tag

Formal product releases use a SemVer tag:

```text
vMAJOR.MINOR.PATCH
```

Examples:

```text
v5.0.0
v5.1.0
v5.0.1
```

SemVer is the public release identity. It is not stored in `VERSION`; it is represented by an annotated Git tag and GitHub Release.

Semantics:

| Change | Version |
|---|---|
| Breaking product, skill layout, RPC contract, output contract, or removal of public entry points | major |
| Backward-compatible new domain, tool, RPC, report, or product capability | minor |
| Bug fix, docs, small correction without breaking contracts | patch |

## 3. v5.0.0 baseline

`v5.0.0` is the first unified product release using the single-skill architecture.

Baseline decision:

- Product release: `v5.0.0`
- Runtime CalVer: `2026.09.16.0`
- Baseline commit: `6f407df`
- Release theme: unified `liki` skill, ready-to-use payload contracts, advisory safety, and cross-domain deterministic golden tests.

Breaking scope included:

- removal of the four old skill entry points;
- consolidation into one installable `liki` skill;
- unified `ENTRY.md` routing;
- `TOOLS.md` / `RPC.md` payload contracts;
- replacement of divination hard blocking with `safety_advisory`;
- explicit fail-closed discover scope and required-method checks.

## 4. Release identity

Use both identifiers in release material:

```text
Liki v5.0.0
Runtime CalVer: 2026.09.16.0
Commit: 6f407df
```

Recommended Git command:

```bash
git tag -a v5.0.0 6f407df -m "Liki 5.0.0 — Unified Chinese Metaphysics Skill Platform"
git push origin v5.0.0
```

The GitHub Release should use the same tag and include the CalVer and commit in its notes.

## 5. Release decision rules

A normal documentation or bug-fix push does not automatically create a SemVer release.

Create `vX.Y.Z` only when the main branch is green and the milestone is intentionally released.

- Breaking public behaviour, payload contract, skill layout, RPC removal, or old-entry removal → major.
- New domain, tool, RPC, report, or backward-compatible capability → minor.
- Fixes or corrections without breaking contracts → patch.

If breaking work lands on main after `v5.0.0`, the next release must be `v6.0.0`; do not hide breaking changes in `v5.1.0`.

## 6. Release process

1. Confirm main is green.
2. Confirm `skills/liki/VERSION` and `engine/cmd/liki/VERSION` are the intended runtime CalVer.
3. Confirm `CHANGELOG.md` contains the milestone entries.
4. Create the annotated SemVer tag on the approved commit.
5. Push the tag.
6. Create the GitHub Release with both versions and release notes.
7. Verify release artifacts and CI.
