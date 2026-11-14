# Liki tests

`make test` runs the deterministic Python and Go suites. Release packaging has
its own content gate.

## Layout

| Suite | Scope |
| --- | --- |
| `tests/` | Domain semantics, release structure, deployment generation, packaging |
| `counsel/tests/` | Counsel tool models, factors, assertions, naming, divination |
| `engine/**/*_test.go` | Engine algorithms and MCP wire contracts |
| `mcp/tests/` | Product gateway and payment flow |

## Commands

```bash
make test          # Python domain suite, engine tests, engine MCP smoke
make test-content  # Python domain/feature suite
make test-engine   # Go engine tests
make test-counsel  # Counsel integration suite with a local engine
make check-content # Build and verify distribution packages
```

When changing counsel dependencies, update both `counsel/pyproject.toml` and
`counsel/requirements.txt`, then regenerate the lock:

```bash
uv pip compile --universal --python 3.12 --generate-hashes \
  counsel/requirements.txt -o counsel/requirements.lock
```

## Placement

- Engine algorithm or wire behavior: `engine/internal/...`
- Counsel tool behavior: `counsel/tests/`
- Skill archive/deployment behavior: `tests/`
- MCP gateway/paywall behavior: `mcp/tests/`

Keep tests focused on behavior that a contributor or release can observe. Avoid
source-text-only assertions unless they protect safety, licensing, or release
integrity.
