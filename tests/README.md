# Liki test model

This repository has deterministic tests and curated data assets. Do not treat functional smoke as proof of metaphysical accuracy.

## Current test layers

| Layer | Entry | Scope |
| --- | --- | --- |
| Content suite | `make test-content` | README, skill layout, tool manifests, factors, assertions, schemas, data flow |
| Python lint | `make lint-python` | Ruff E/F/W/I over counsel and active scripts |
| Python vulnerabilities | CI `pip-audit` | Known CVEs in the hashed counsel dependency lock |
| Engine units/integration | `make test-engine` | Go domain packages, agent registry, HTTP middleware, integration tags |
| Toolchain gate | `scripts/check-go-version.sh` | Rejects Go toolchains older than 1.26.9 |
| Engine vulnerabilities | CI `govulncheck ./...` | Reachable Go standard-library/module vulnerabilities |
| Engine MCP smoke | `make test-engine-mcp` | Internal `/mcp` root and `/mcp/{domain}` surfaces, `server/discover`, `tools/list`, `tools/call` |
| Counsel MCP integration | `make test-counsel` | Bootstraps a locked Python environment and local engine, then tests all counsel tools |
| Assertion data check | `python3 scripts/check_factor_data.py` | Factor/assertion table integrity, provenance, reachability inputs |
| Documentation check | `python3 scripts/check_docs.py skills/liki` | Distributed skill documentation references |
| Markdown lint | `make lint-md` | Locked `markdownlint-cli2` over all skill markdown |
| Full 160-question data check | `make full-data` | Rule coverage and zero-hit detection; not answer grading (starts a local engine-mcp) |
| Golden suites | `make golden` | Deterministic multi-source calibration fixtures already consumed by engine tests |
| Distribution build | `make build-archive` | Self-contained skill archive and expert methodology cards |

Regenerate the counsel dependency lock after changing `counsel/requirements.txt`:

```bash
uv pip compile --universal --python 3.12 --generate-hashes \
  counsel/requirements.txt -o counsel/requirements.lock
```

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
