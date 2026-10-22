# MCP contract model

## Goal

MCP tools use one source of truth per tool:

```text
typed wire DTO → JSON Schema → runtime validation → structuredContent
```

The tool description and DTO field descriptions are part of the public MCP
contract. Internal domain structs are not exposed directly: each tool defines
an explicit wire input and output. This prevents accidental leakage of private
runtime shapes and keeps generated schemas stable.

### Time fields

Tools that consume a precise time use `instant` plus an explicit `timezone`:

```json
{
  "instant": "1990-05-20T12:00:00+08:00",
  "timezone": "Asia/Shanghai"
}
```

`instant` is an absolute RFC3339 instant. Canonical output is UTC and ends
with `Z`; numeric input offsets are accepted and normalized immediately.
`timezone` must be an explicit IANA identifier. Offset is never used to infer
timezone, and missing timezone or longitude fails closed when the tool needs
them.

## Go pattern

`engine/internal/mcpcontract` owns the typed MCP registry. A tool registers:

```go
type ExampleInput struct {
    City string `json:"city" jsonschema:"City name to resolve."`
}

type ExampleOutput struct {
    Latitude float64 `json:"latitude" jsonschema:"Latitude in decimal degrees."`
}
```

The registry derives both JSON Schemas, validates arguments and the returned
output, and returns the output DTO as MCP `structuredContent`. Legacy RPC
envelopes are not part of the MCP result.

Rules:

- every wire DTO field has a `jsonschema` description;
- object DTOs close unknown fields with `additionalProperties: false`;
- a tool that replaced a legacy raw-JSON method sets
  `Tool.TolerateUnknownArgs`, because the legacy params schemas left
  `additionalProperties` unset and therefore ignored unknown arguments;
- output fields are required unless explicitly modeled as optional;
- tools are registered in deterministic order;
- business failures become MCP tool errors, not private RPC envelopes;
- a migrated tool is removed from the legacy raw-JSON registry in the same
  change.

## Migration

The migration is complete: all 27 engine tools are typed contracts and the
legacy raw-JSON registry (the `RPCRegistry` type, the `tools_*.go` schemas and
the `{"_product","data"}` envelope) has been retired.
`engine/internal/mcpcontract` is the single source of truth: `engine.go` holds
the tool catalogue (`NewEngine()`, the registration point, and the wire
vocabulary the schemas derive from), `registry.go` holds the registry mechanism,
and the wire DTOs are ordinary source maintained alongside their handlers.

The bazi/ziwei/qimen DTOs were derived once from the legacy schemas plus the
captured engine payloads committed under
`engine/internal/mcpcontract/testdata/actual/`. That generator was deliberately
not retained: the DTOs are the contract now. `TestEngineContractWireFidelity`
replays all 67 captured payloads through the typed registry and deep-compares
the JSON, so a DTO edit that drops, renames, or reshapes a field fails.

Five contract limits to keep in mind when reviewing the migrated tools:

- the legacy qimen result schema used `oneOf` branches conditioned on the
  requested method (a rotating chart carries `tian_qin`, only 鸣法 allows
  `中门`/`an_gan_zhi`, and so on). The typed DTO merges those branches into one
  object, so the output schema is a superset and those conditional constraints
  are no longer enforced;
- qimen.chart's params schema collapsed the legacy `allOf` branches that
  conditioned `quarter_rule` on `scope=quarter` and `dingju_method`/
  `school=mingfa_feipan` on their scopes. The handler still rejects every
  illegal combination, and `TestMigratedQimenMethodMatrix` pins that; only the
  advertised input schema is wider. `yong_shen` keeps the legacy
  `minItems: 1`/`uniqueItems: true` through the `QimenYongShenSymbols` type, and
  `birth_date` is a plain string (the legacy `date`/`date-time` format hint is
  annotation-only, and `qimen.ParseBirthDate` still decides);
- output `required` is taken from the legacy `required` lists intersected with
  the captured payloads. A field the engine omits through `omitempty` is
  therefore modelled as optional even when the legacy schema declared it
  required;
- `ElementName` also accepts the empty string. The engine documents the empty
  value for bazi 调候 `TiaoHouResult.SecondaryWuxing` (喜神 may be absent,
  `engine/internal/engine/bazi/bazi_correctness_test.go`), and a strict enum
  would reject real charts. `TestEngineContractBaziTiaoHouAllowsEmptySecondaryWuxing`
  pins that case;
- migrated legacy tools accept unknown arguments, because the legacy registry
  validated params against schemas that left `additionalProperties` unset
  (`bazi.chart` tolerates `tianwen.time`'s `longitude`, and `ziwei.daxian`
  accepts a `ziwei.fullchart` payload). Strictness is kept where the legacy
  schema declared it: `qimen.chart`'s root object and the ziwei flow tools'
  root/`target_lunar` objects stay closed through `Tool.KeepClosedInputPaths`.
  Tools written directly against the typed registry (`time.now`, and the rest
  of the non-legacy tools) still reject unknown arguments.

Array-root outputs preserve their wire shape. The typed registry still
validates every element, but MCP `outputSchema` remains omitted for non-object
roots, matching the current engine MCP conformance surface.

## Python pattern

`counsel/app/contracts` owns the typed counsel tool surface. A tool declares its
models and the registry projects everything from them:

```python
class ExampleInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    city: Annotated[str, Field(description="City name to resolve.")] = Field(
        description="City name to resolve.",
    )

class ExampleOutput(BaseModel):
    model_config = ConfigDict(extra="allow")
    latitude: Annotated[float, Field(description="Latitude in decimal degrees.")] = Field(
        description="Latitude in decimal degrees.",
    )

TOOL_EXAMPLE = ToolContract(
    name="example",
    description="…",
    input=ExampleInput,
    output=ExampleOutput,
)
```

`ToolContract.input_schema` / `.output_schema` are emitted by Pydantic's own
generator (a `GenerateJsonSchema` subclass). `app.mcp_tool_adapter` registers the
contract on the MCP SDK, which owns the rest: the input model validates the
arguments before the handler runs, the output model validates the result before
it is published, and the result dict is what the client receives as
`structuredContent`. No JSON Schema is written or parsed by hand, and the runtime
never reads a manifest file.

Rules:

- every tool field carries a description (input and output alike); `make check`
  enforces it for required fields and enum fields through the schema budget lint;
- `extra="forbid"` closes input objects (`additionalProperties: false`), while
  output objects stay open so new fields do not fail old clients;
- `unique_list[T]` restores `uniqueItems: true`, which Pydantic does not project
  on its own;
- `$defs` are inlined, so a published schema stays a flat document;
- the cross-field rules the legacy manifests declared as `allOf`/`if`-`then`
  live in `@model_validator(mode="after")`: Pydantic cannot project them into
  the schema, and the schema is not the only contract anyway, so the boundary
  keeps rejecting those combinations.

There is no committed schema file. Every consumer imports the models:
`scripts/check_mcp_schema.py` projects the published schemas from
`app.contracts` for the budget and description lint, and the root test suite
(`tests/`) imports them where it asserts on the tool surface, so a contract edit
can never leave a stale copy behind. A client that needs the schemas asks the
server: `tools/list` is the published surface.

The one committed derived artifact is the tool-name catalog,
`tools/mcp-tool-catalog.json`. It holds names only, never schemas, and it is
generated by `python3 scripts/check_mcp_schema.py --emit` from the typed engine
registry plus `app.contracts.TOOL_GROUPS`; `make check` runs the same script in
`--check` mode and fails on any drift, so the file is never hand-edited.

The Go engine suite cannot import Python models, so the one cross-language
consumer goes through a bridge instead of a copy: `app.contracts.tool_vocabulary()`
publishes each skill group's tool names, property names, and enum values, and
`engine/internal/mcpcontract/skill_docs_sync_test.go` runs
`scripts/counsel_tool_vocabulary.py` to load them. That keeps the skill-document
field contract test resolving counsel references against the models; the engine
job installs `pydantic` for it, and `COUNSEL_PYTHON` selects the interpreter when
the default one does not own the counsel dependencies. The check fails rather
than skips when the bridge cannot run, so an unreadable source of truth is never
a silent pass.

## Counsel migration

All 13 counsel tools (`bazi_yongshen`, `compute_factors`, `natal_query`,
`period_query`, `naming_surname`, `naming_pick`, `naming_char`, `naming_compose`,
`naming_check`, `liuyao_snapshot`, `liuyao_query`, `qimen_snapshot`,
`qimen_query`) are typed contracts; the runtime no longer reads a JSON manifest.
`bazi_yongshen` is published by the bazi server and the aggregate server; the
ziwei server does not advertise a bazi-only capability.

The yongshen decision has one typed wire model shared by every tool that carries
it. `bazi_yongshen(chart)` takes a typed engine `bazi_chart` object and returns a
`SelectedYongShen`: `selected` (`yong/xi/ji/source/confidence/known_fields`),
resolution `status`, conflicts, `rule_id`, policy and rule provenance, the
`ge_ju_context`, branch `evidence`, and the two SHA-256 digests. `evidence`
includes the 从格 effective/rooted counts used to verify or reject the candidate,
so a `needs_review` result is auditable over MCP. The complete pre-factor snapshot
and alternatives stay in the `compute_factors` envelope, not inside the decision
object. The decision digest binds the source chart, the published decision, and
its evidence; it is not a digest of the internal pre-factor snapshot. For a bazi
chart, `compute_factors` embeds the same decision at
`factors.selected_yongshen`, produced by the same resolver; a ziwei factor
snapshot does not contain that element. `naming_check` accepts the
`ConsumedYongShen` subset of that object and returns the `YongShenSource` subset
it consumed. `SelectedYongShen`, `ConsumedYongShen`, and `YongShenSource` keep
`additionalProperties` open so a policy can add fields without breaking older
clients; the stable decision fields themselves are enums, required fields, or
digest patterns.
The `naming_check` fields `yongshen`/`xishen`/`jishen` remain an explicit
five-element strategy for callers without a full birth chart, and are mutually
exclusive with `selected_yongshen`.

The hand-written manifests are gone from the runtime; they were frozen once
under `counsel/tests/testdata/legacy-tool-manifests/` as the migration golden, and
`counsel/tests/test_contract_fidelity.py` proves the migration:

- the published `inputSchema`/`outputSchema` are equal to the frozen manifest
  schemas after canonicalisation (nullable spelling, implicit
  `additionalProperties`, empty `items`, annotation-only `title`/`default`).
  The frozen manifests predate `bazi_yongshen`, which is a new tool rather than
  a migrated one, and the yongshen fields upgraded from free-form objects. Both
  exclusions are asserted against the typed schemas separately;
- every clause the models dropped from the schema is enforced at runtime by a
  model validator: on a corpus that walks each conditional branch, no call the
  contract accepts was rejected by the legacy schema;
- two qimen clauses in the legacy manifest were written without their
  `properties` keyword, so JSON Schema silently ignored them
  (`scope=quarter`/`scope=month|year` with the hour-scope method fields). The
  models enforce the documented intent, which the engine catalog rejects as
  well, and the test pins those cases as intentionally stricter.
