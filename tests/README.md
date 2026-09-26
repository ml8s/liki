# Liki test model

This repository has deterministic contract tests and model-backed evaluation assets. Do not treat a model benchmark as a functional regression suite, and do not treat functional smoke as proof of metaphysical accuracy.

## Current test layers

| Layer | Entry | Scope |
| --- | --- | --- |
| Skill contracts | `make test-contracts` | README, skill layout, tool manifests, factors, assertions, schemas, data flow |
| Python lint | `make lint-python` | Ruff E/F/W/I over counsel and active scripts |
| Python vulnerabilities | CI `pip-audit` | Known CVEs in the hashed counsel dependency lock |
| Engine units/integration | `make test-engine` | Go domain packages, agent registry, HTTP middleware, integration tags |
| Toolchain gate | `scripts/check-go-version.sh` | Rejects Go toolchains older than 1.26.6 |
| Engine vulnerabilities | CI `govulncheck ./...` | Reachable Go standard-library/module vulnerabilities |
| Engine MCP smoke | `make test-engine-mcp` | Internal `/mcp` root and `/mcp/{domain}` surfaces, `server/discover`, `tools/list`, `tools/call` |
| Counsel MCP integration | `make test-counsel` | Bootstraps a locked Python environment and local engine, then tests all counsel tools |
| Assertion data check | `python3 scripts/check_schema.py` | Factor/assertion table integrity, provenance, reachability inputs |
| Documentation check | `python3 scripts/check_docs.py skills/liki` | Distributed skill documentation references |
| Markdown lint | `make lint-md` | Locked `markdownlint-cli2` over all skill markdown |
| Node vulnerabilities | CI `npm audit` | Known advisories in the locked markdown lint toolchain |
| Full 160-question data check | `LIKI_MCP_URL=http://engine-mcp:8081/mcp python3 scripts/eval_hybrid.py` | Rule coverage and zero-hit detection; not answer grading |
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

## Route contract

The public paths are `/engine/mcp/{domain}` and `/counsel/mcp/{domain}`. Caddy strips `/engine` or `/counsel`; the services themselves see `/mcp/{domain}`. Tests must not document or assert service-level public prefixes.

## Model-backed assets

`tests/benchmark/mingli160/` and `tests/skillup/` contain historical model evaluation assets. Their runners predate the unified MCP skill surface and are not part of `make gate`. Keep answer data isolated if they are revived; migrate the runner to engine-mcp before using them as release evidence.
