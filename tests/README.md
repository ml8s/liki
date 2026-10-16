# Liki test model

This repository has deterministic tests and curated data assets. Do not treat functional smoke as proof of metaphysical accuracy.

## Current test layers

| Layer | Entry | Scope |
| --- | --- | --- |
| Content suite | `make test-content` | README, skill layout, tool manifests, factors, assertions, schemas, data flow |
| Python lint | `make lint-python` | Ruff E/F/W/I over `counsel/`, `scripts/`, and root `tests/` |
| Python vulnerabilities | CI `pip-audit` | Known CVEs in the hashed counsel dependency lock |
| Engine units/integration | `make test-engine` | Go domain packages, MCP contract registry (`internal/mcpcontract`), HTTP middleware, integration tags |
| Toolchain gate | `scripts/check-go-version.sh` | Rejects Go toolchains older than 1.26.9 |
| Engine vulnerabilities | CI `govulncheck ./...` | Reachable Go standard-library/module vulnerabilities |
| Engine MCP smoke | `make test-engine-mcp` | Internal `/mcp` root and `/mcp/{domain}` surfaces, `server/discover`, `tools/list`, `tools/call` |
| Counsel MCP integration | `make test-counsel` | Bootstraps a locked Python environment and local engine, then tests all counsel tools |
| Assertion data check | `python3 scripts/check_factor_data.py` | Factor/assertion table integrity, provenance, reachability inputs |
| Documentation check | `python3 scripts/check_docs.py skills/liki` | Distributed skill documentation references |
| Documentation path check | `python3 scripts/check_doc_paths.py` | Every Markdown link target and top-level code-span path in the repository resolves |
| Markdown lint | `make lint-md` | Locked `markdownlint-cli2` over all skill markdown |
| Full 160-question data check | `make full-data` | Rule coverage and zero-hit detection; not answer grading (starts a local engine-mcp) |
| Golden suites | `make golden` | Deterministic multi-source calibration fixtures already consumed by engine tests |
| Distribution build | `make build-archive` | Self-contained skill archive and expert methodology cards |

## Test placement

Put a test where its contract lives; the suite that owns the contract also owns the test.

| Contract under test | Location | Entry |
| --- | --- | --- |
| Engine tool schema, DTO, wire fidelity | `engine/internal/mcpcontract/*_test.go` | `make test-engine` |
| Engine domain algorithm or data table | `engine/internal/engine/<domain>/*_test.go` | `make test-engine` |
| Counsel tool models, factors, assertions | `counsel/tests/` | `make test-counsel` |
| Cross-language surface (skill docs, both registries) | `engine/internal/mcpcontract/skill_docs_sync_test.go` | `make test-engine` |
| Repository surface (README, skill layout, deployment, versions, CI contracts) | root `tests/` | `make test-content` |
| Independent domain truth (data only, never implementation-derived) | `tests/fixtures/domain_oracle/` | oracle tests |
| Characterization snapshot (drift detector) | `tests/golden/<domain>/` | `make golden` |

Notes:

- Root `tests/` is a package (`tests/__init__.py`), so its modules cannot collide with
  same-named modules under `counsel/tests/`; the counsel suite has its own rootdir.
- Golden generators write into `engine/internal/engine/*/testdata/` and resolve that path
  from the repository root, so they run from any cwd. Never hand-edit a golden; change the
  generator and regenerate. See `tests/golden/README.md`.

## Static check scripts

`make check` runs no tests and writes nothing: it is the static half of the push gate. Its
script set is pinned by `tests/test_check_scripts_contract.py`, so adding a script to the
`check` recipe without listing it here (or leaving a listed script unwired) fails.

| Script | Contract |
| --- | --- |
| `scripts/lint-md.sh` | Locked `markdownlint-cli2` over the declared Markdown globs |
| `scripts/lint-python.sh` | Locked Ruff over `counsel/`, `scripts/`, and root `tests/` |
| `scripts/check-go-version.sh` | Rejects Go toolchains older than 1.26.9 |
| `scripts/check-runtime-versions.sh` | Local node/python must equal `.nvmrc`/`.python-version` |
| `scripts/check_actions_pinned.py` | Rejects mutable workflow/action references |
| `scripts/check_factor_data.py` | Factor/assertion table integrity, provenance, reachability inputs |
| `scripts/check_docs.py` | Distributed skill documentation: assertions, paths, budgets, method whitelist |
| `scripts/check_doc_paths.py` | Markdown link targets and top-level code-span paths resolve |
| `scripts/check_workflow.py` | CI workflow structure and job wiring |
| `scripts/check_mcp_schema.py` | MCP tool schemas, context budgets, and `tools/mcp-tool-catalog.json` drift (`--emit` regenerates it) |
| `scripts/check_expert_docs.py` | Expert-pack methodology cards against the root skill |
| `scripts/agent_runtime_pin.py` | Agent runtime image pin is internally consistent |
| `scripts/check-exec-bits.sh` | Git executable bits for scripts invoked without an interpreter |
| `scripts/sync_expert_packs.py` | Expert methodology cards are in sync with the root skill |
| `scripts/check_workbuddy_connectors.py` | WorkBuddy connector packages: metadata, single MCP server, HTTPS, version gate, credentials, path-only icon glyphs |

The counsel direct dependencies are declared twice on purpose --
`counsel/pyproject.toml [project.dependencies]` is the metadata surface and
`counsel/requirements.txt` is the lock input -- so change both and let
`tests/test_counsel_dependency_parity.py` confirm they agree, then regenerate the lock:

```bash
uv pip compile --universal --python 3.12 --generate-hashes \
  counsel/requirements.txt -o counsel/requirements.lock
```

The lock is frozen between deliberate upgrades; the same command today would also bump
transitive pins, so regenerating it is its own reviewed change.

The full push gate is:

```bash
make gate
```

## Route boundary

The public paths are `/engine/mcp/{domain}` and `/counsel/mcp/{domain}`. Caddy strips `/engine` or `/counsel`; the services themselves see `/mcp/{domain}`. Tests must not document or assert service-level public prefixes.

## Oracle & golden policy

Independence matters more than volume. Keep the three tiers distinct:

- **Independent truth** — `tests/fixtures/domain_oracle/*.json`: hand-curated hard facts with
  `basis` / `sources`, explicitly **not** generated from the implementation. Go/Python oracle
  tests read only this directory; never copy implementation internals as `expected`, and never
  reverse-edit `expected` to make a failing implementation pass (see
  `tests/fixtures/domain_oracle/README.md`).
- **Regression guards** — `tests/golden/*`: deterministic characterization snapshots. They detect
  *drift*, not correctness; a golden that is regenerated after a change proves nothing on its own.
  Where an external source exists (e.g. bazi/黄历 multi-oracle consensus), bind the golden to it;
  otherwise treat it as a change-detector only.
- **Coverage dataset** — `tests/mingli160/`: 32 birth records and their question
  grouping. The set informed the natal assertion table during development, so it
  is not held out and must not be reported as independent accuracy. A real
  accuracy claim needs a frozen assertion version and a second, tuning-untouched
  question set.
