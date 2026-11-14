# Contributing to Liki Skill

> 简体中文贡献指南：[CONTRIBUTING.zh.md](./CONTRIBUTING.zh.md)

Thanks for considering a contribution to Liki. This guide covers how to report
issues, set up a development environment, and land changes safely. The project
is enforced by deterministic gates — most conventions described here are pinned
by scripts and tests, so CI tells you when something drifts.

## Reporting issues

- **Inaccurate astrology results** — include birth date/time, birthplace, the
  expected result, and the actual result.
- **Unexpected Skill behavior** — describe which flow was triggered, what you
  expected, and what actually happened.
- **New capability proposals** — describe the user scenario as concretely as
  possible.

Suspected security vulnerabilities go through the private channel in
[SECURITY.md](./SECURITY.md), not public issues.

## Development environment

Toolchain versions are declared in exactly one place; local and CI read the
same source, and `make check` verifies they agree:

| Tool | Version | Declared in | Verified by |
| --- | --- | --- | --- |
| Go | 1.26.11 | `go.mod` + `GOTOOLCHAIN` | `make go-version-check` |
| Python | 3.12.14 | `.python-version` | `make runtime-version-check` |
| Node | 24.14.1 | `.nvmrc` | `make runtime-version-check` |
| golangci-lint | v2.13.2 | `deps/golangci-lint.version` | reconciliation test |
| ruff | 0.16.0 | `RUFF_VERSION` in `scripts/lint-python.sh` | — |

First-time setup:

```bash
make hooks        # point core.hooksPath at .githooks (pre-commit / pre-push gates)
npm ci            # locked local Markdown lint dependency (markdownlint-cli2)
make python-env   # install the exact .python-version interpreter + counsel venv via uv
```

Notes:

- GitHub Actions are pinned to 40-character commit SHAs. Dependabot bumps them
  in one grouped weekly PR; read the release notes before merging — major
  versions often break (e.g. `docker/login-action` v4).
- Engine lint uses golangci-lint v2 (config: `engine/.golangci.yml`). Install
  it with the official binary script, **not** `go install` (golangci-lint is
  tightly coupled to Go versions and upstream discourages that method):

  ```bash
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin
  ```

- Python linting uses ruff 0.16.0 (`make lint-python`); vulnerability scanning
  uses `govulncheck` (Go) and `pip-audit` (counsel hashed lock), run separately
  in CI.
- On mainland-China networks, `export UV_DEFAULT_INDEX=https://pypi.tuna.tsinghua.edu.cn/simple`
  speeds up uv.

## Branch model

`main` is the only long-lived branch — no `dev`, no `master`, no long-lived
release branches. Maintainers land small, focused commits directly on `main`.
External contributors fork and use short-lived branches; PRs still target
`main`.

A release means: build images locally pinned to each repo's commit sha →
`liki-deploy` runs full-system e2e → deploy and smoke. CI never builds images.

## Making changes

1. Sync latest `main`: `git pull --ff-only`.
2. Install git hooks once: `make hooks`.
3. Update `CHANGELOG.md` when behavior is user-visible (and README stats when
   they change).
4. README / user-guide changes must preserve the existing heading structure
   and the bilingual contract; run `make check`.
5. Keep unrelated changes in separate commits, and describe what changed and
   why.

Commit messages follow conventional commits (`feat`, `fix`, `refactor`,
`docs`, `chore` with an optional scope), matching the existing history.

## Code and documentation standards

### Language and service boundaries

- `engine-mcp` performs deterministic charting, calendar, and feng shui
  computation only; `counsel-mcp` handles factors, assertions, naming, and
  divination judgment only. The edge owns the public `/engine` and `/counsel`
  prefixes; inside the services only `/mcp` is allowed. Domain documentation
  describes capabilities and variable bindings — never hand-write transport
  details.
- The root `skills/liki/SKILL.md` is product routing only; each domain is
  entered through its `ENTRY.md`. SKILL.md is deliberately Chinese-first with
  original terms preserved — that is a product language decision, independent
  from this repository's documentation language.

### Skill documentation conventions

Pinned by `scripts/check_docs.py` and `tests/test_unified_skill_structure.py`:

- Domain `ENTRY.md` files carry in-domain routing, tool contracts, and hard
  boundaries only. App cards are flat: `references/<domain>-app-<goal>.md`
  (e.g. `natal-app-marriage.md`), carrying flows, interaction gates, and output
  templates. Methodology cards are flat:
  `references/<domain>-<system>-<card>.md` (e.g. `natal-bazi-geju.md`).
- Every piece of information is defined in exactly one place; routing
  reachability and reference shapes are test-pinned.
- Methodology cards carry no frontmatter — identity comes from flat naming and
  the ENTRY routing tables; there is no consumer for card metadata.
- `## 输出模板` sections must contain a real template body (code block / list
  / table); examples go under `### 示例`.
- One rule, one definition: domain-wide policies live in that domain's
  `ENTRY.md` / `TOOLS.md`; cross-system synthesis lives in `app/` cards. Where
  channel-pack self-containment forces a same-source duplicate (such as the
  yongshen priority order shared by natal / naming), tests pin the copies as
  byte-identical.
- Cards reference tool names and tool field names only — never internal data
  files (`data/*.json`).
- Cross-card references always use the flat full path
  `references/<flat-name>.md`; bare names, legacy nesting, cross-domain, root
  shared cards, and `scripts/` references fail closed in the generator.
- Prefer positive guidance: procedural docs say what to do and how; domain
  prohibitions distilled from real errors (grade-skip prevention, character
  pool filters, anti-repetition red lines) stay as they are.

### Single expert and channel assets

- `skills/liki/` is the only content source; `distribution/workbuddy/liki-pro/`
  holds WorkBuddy channel assets only.
- `distribution/liki-expert/deployment.yaml` and
  `distribution/liki-expert/profile.json` are the only deployment configs.
- The single handwritten identity source is `skills/liki/SKILL.md`; channel
  entry files are generated from it plus
  `distribution/workbuddy/liki-pro/channel.json` and the paid overlay. The
  `make check` drift gate rejects a second identity copy inside packages.

### Intentional duplication (do not merge)

Reviewers will spot the following as mergeable — each is deliberate. Confirm
the reason still holds before touching any of them.

| Duplication | Why it stays |
| --- | --- |
| `engine/**/zz_load_test.go` (12 files) | Go cannot share `TestMain` across packages; extracting a common package would pull `testing` into non-test build dependencies. 9 call their own `Load()`, 3 call a dependency package's `Load()` (fengshui→luoshu, tianwen→ganzhi, http→tables). |
| Script preamble `set -eo pipefail` + `ROOT="$(cd "$(dirname "$0")/.." && pwd)"` (7 scripts) | Every script must run standalone (`make`, CI, manual invocation); a shared include would introduce a sourcing dependency. |
| Empty `tests/__init__.py` | Root `tests/` is an explicit package so pytest import names are `tests.<module>`; `counsel/tests/` has no `__init__.py` and uses bare basenames, so the two suites never shadow each other. |
| `skills/liki/SKILL.md` body duplicated in channel packages | The root file is the only handwritten identity source; channel entries are rendered from it per platform (WorkBuddy adds a methodology card index). Same-source keeps the marketplace entry, ADK skill, and deployment instruction from drifting (`make check` drift gate). |
| Single-file directory `tools/` | Contains only the generated `mcp-tool-catalog.json`: kept as `scripts/check_mcp_schema.py --emit` output with `make check` drift detection, not merged or hand-edited. |

### Design principles (why it is built this way)

- **Historical events validate the whole framework, not a single variable.**
  Calibration and conclusion verification fall back to the combined
  structure + yongshen + luck-cycle reading framework, never to a single
  yongshen choice — an event is the synthesized outcome of the whole frame and
  cannot be back-solved from one variable (lesson from v1.23.0). Enforced in
  `references/natal-app-mingshu.md` and `references/natal-bazi-calibration.md`.
- **Yongshen has exactly two schools (扶抑 / 调候) and must aggregate to one
  conclusion.** The structure (格神) is not the yongshen; it only participates
  in the aggregation decision as structural evidence
  (`references/natal-bazi-yongshen.md`). Never list candidates and ask the user
  to choose — that delegates professional judgment to the user (lesson from
  v1.16.0, now enforced by the yongshen aggregation decision table).
- **MCP call boundaries are layered.** See "Language and service boundaries"
  above; this is an architectural rule, not a convention.

## Testing and validation

| Command | What it does |
| --- | --- |
| `make check` | Static checks: lint (md/python/go fmt), versions, pinned actions, channel package determinism, docs, domain language. No tests. |
| `make test` | All deterministic tests: Python contract suite + engine coverage floor + MCP protocol smoke. |
| `make gate` | Full source push gate — the same checks and tests as CI (lint-engine, python deps, check, city data, archive + deployment builds, full test suite). |

Both git hooks run these automatically: pre-commit lints staged Markdown;
pre-push runs `make gate`. Expect the push gate to take several minutes.

## Release version model

- Internal runtimes carry no release version; build identity is the Git short
  SHA, and image tags and OCI revisions share that source.
- No hand-maintained version files, no `make version`, no internal version
  reconciliation.
- If WorkBuddy / SkillHub platforms force semver,
  `scripts/version_projection.py` derives it from the Git commit time at
  packaging time. That is an external-platform compatibility layer, not an
  internal release contract.

## Pre-push checklist

**Renaming a method or function** — search every reference site:

```bash
# code and scripts
grep -rn "old-name" --include="*.go" --include="*.sh" --include="*.py"
# skill documentation
grep -rn "old-name" skills/liki/references/*-app-*.md skills/liki/references/*-ENTRY.md skills/liki/SKILL.md
```

**Renaming a domain concept or field** — the single naming source and alias
gate:

```bash
python3 scripts/sync_domain_language.py
python3 scripts/check_domain_language.py
```

Do not invent synonyms in code, contracts, comments, tests, or docs. If a
legacy field must survive, register a compatibility `scope`, the reason, and a
removal version in `tools/domain-language.json`; the checker allows the old
name only in those places.

**Adding a new MCP tool** — the Pydantic / typed registry contracts are the
single source of truth; add contracts and tests in the same change:

```bash
# engine: add handler, InputSchema/OutputSchema, and tool schema tests
grep -R "bazi_chart" engine/cmd/engine-mcp engine/internal/mcpcontract
# counsel: add the Pydantic contract in app/contracts, then implementation and error-mapping tests
grep -R "natal_compute_factors" counsel/app counsel/tests
```

**Changed skill documentation and a package build is needed:**

```bash
make build-archive
```

**Full local CI before pushing:**

```bash
make gate
```

## Related documents

- [docs/SYSTEM.md](./docs/SYSTEM.md) — system overview: four-repo split, authoritative sources, expert definition structure, MCP connector model, generation/verification chains, change propagation.
- [docs/MCP_CONTRACTS.md](./docs/MCP_CONTRACTS.md) — MCP contract model: typed wire DTOs → JSON Schema → runtime validation, on both engine and counsel.
- [distribution/workbuddy/README.md](./distribution/workbuddy/README.md) — WorkBuddy skill/expert package structure, artifacts, and verification flow.
- [docs/README_REVIEW.md](./docs/README_REVIEW.md) — README review checklist (brand, SEO, GEO, facts).
