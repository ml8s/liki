# Performance model

This page records performance policy for the MCP runtime and benchmark entry
points:

```bash
make bench-engine
make bench-engine-save BASE=before
make bench-engine-compare OLD=before NEW=after
make bench-engine-1cpu
```

## Engine policy

- Untrusted MCP arguments are validated on every call.
- Simple generated input schemas use a bounded fast validator. Advanced
  composition, references, regexes, and similar keywords automatically fall
  back to the generic JSON Schema validator; fixture mutations check that both
  paths produce the same accept/reject result.
- Large-domain JSON that is already locked by wire-fidelity fixtures can use a
  typed raw projection instead of a domain→DTO→JSON round trip.
- Successful tool results default to structured-only output. `compat` mode is
  retained for legacy clients that still read duplicate `TextContent`.
- Published output schemas are checked against actual JSON by contract tests.
- Production calls do not perform a reflective output-schema validation pass.
- Tool registration is lazy, and all schemas are materialized concurrently
  before the process starts serving; contract drift fails startup rather than
  the first request.
- Pure calendar values may use bounded caches; user charts are not cached.
- Behavior-preserving optimizations need golden/oracle and wire-contract tests.
- Fused operations are expressed as explicit pure projections. Tool inputs are
  immutable artifacts, and a projection returns the exact facts its composition
  stage needs instead of mutating request state.

## Counsel optimizations (implemented)

| Area | What was done | Key pattern |
| --- | --- | --- |
| Truth-table evaluator | AOT compiler: rows → FactorPlan (pre-parsed expressions, topo-sorted, single pass) | Precompute static structure, evaluate per-call |
| Ziwei palace facts | PalaceIndex: O(1) hash lookup replaces O(n) linear scan per operator call | Build index once, query many times |
| Series batch RPC | `count` param on 10 time-series tools; counsel batches per side in chunks bounded by engine `MaxSeriesCount` | Reduce 42 round-trips to 2 for a typical ≤50-year range |
| Ziwei chart decode | `runSeriesPrepared`: chart decoded once outside loop, not per-year JSON round-trip | Prepare shared state, iterate varying input |
| Counsel assertions | Expected-value validation cached by table identity; redundant `match_rule` validation skipped via `skip_validation` | Eliminate cross-year duplicate checks |
| Natal context | `_TtlCache` for current-year; LRU for natal context by chart digest | Explicit lifecycle, bounded state |
| Period query 21y | 402 ms → ~45 ms (−89 %) | All patterns above combined |

## Follow-ups outside the Go engine

| Area | Status / next step | Guardrail |
| --- | --- | --- |
| counsel engine client | Add one total deadline plus jittered retry budgets | Do not blindly retry random, non-idempotent tools such as auto Liuyao qigua |
| Edge / deployment | Evaluate response compression and CPU-quota scaling | Compare network savings against p95 latency on one-CPU containers |
| MCP SDK | Request quieter stateless lifecycle logging or a debug level hook | Preserve warnings and errors |

## Next engine candidates

1. Generate tool schemas and validators from one source of truth.
2. Collapse fast input validation and typed decode into one schema-driven pass.
3. Generate Qimen and Ziwei JSON encoders from their published wire contracts.
4. Add a production HTTP load budget under the target CPU and memory quota.
