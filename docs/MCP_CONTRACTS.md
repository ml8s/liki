# MCP contract model

The working contract standard is defined in
[MCP_CONTRACT_DESIGN.md](./MCP_CONTRACT_DESIGN.md). Current implementation is the
typed registry described below.

## Goal

MCP tools use one source of truth per tool:

```text
typed wire DTO → JSON Schema → runtime input validation → structuredContent
```

The tool description and DTO field descriptions are part of the public MCP
contract. Internal domain structs are not exposed directly: each tool defines
an explicit wire input and output. This prevents accidental leakage of private
runtime shapes and keeps generated schemas stable.

### Time fields

Birth/event wall-clock input is resolved once by `tianwen.resolve`:

```json
{
  "recorded_date": "1990-05-20",
  "recorded_time": "12:00",
  "record_timezone": "Asia/Shanghai",
  "place_longitude": 116.4
}
```

Chart and flow tools then use only `standard_time` plus longitude:

```json
{
  "standard_time": "1990-05-20T12:00:00+08:00",
  "longitude": 116.4
}
```

`standard_time` is local standard time with no DST, represented as RFC3339 with
a known offset. `Z` is valid UTC standard time; `-00:00` is rejected because
RFC3339 defines it as an unknown local offset. The offset is not an IANA
timezone identity, but it selects the local standard representation used by
date facts. `timezone` is accepted only by `tianwen.resolve`; chart tools fail
closed if it is present. Missing `longitude` fails closed when the tool needs
LAST.

The per-domain projection matrix (solar-term event, standard date, local LAST)
is defined in [TIME_MODEL.md](./TIME_MODEL.md).

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

The registry derives and publishes both JSON Schemas, validates arguments at
the untrusted MCP boundary, and returns the output DTO as MCP
`structuredContent`. Legacy RPC envelopes are not part of the MCP result.
Simple generated schemas use a bounded fast input validator; schemas containing
composition, references, regexes, or other advanced keywords automatically use
the generic JSON Schema validator.
Output JSON is checked against its schema by contract tests through
`Registry.ValidateOutput`; a reflective output validation pass is not repeated
for every production call.
Schemas are materialized concurrently before the HTTP listener starts, so an
invalid derived contract fails process startup instead of a first request.
Successful results are `structuredContent` only: every tool publishes a typed
object-rooted outputSchema, the duplicated text payload was removed as a size
optimization, and all consumers (browser, expert SDK, external agents) read
the structured field. Tool failures still surface as text content with
`isError`.

Tools whose domain JSON shape is already locked by wire-fidelity fixtures may
marshal that domain value directly through `RawTool`; the wire DTO remains the
schema source of truth. This avoids a domain-to-DTO-to-JSON round trip without
publishing an untyped contract.

Rules:

- every wire DTO field has a `jsonschema` description;
- object DTOs close unknown fields with `additionalProperties: false`;
- input schemas validate untrusted MCP arguments on every call;
- time-series `count` inputs use `SeriesCount`, publishing the same
  `[1, MaxSeriesCount]` bounds that runtime validation enforces; omitted count means one item and explicit zero is invalid;
- output schemas validate actual JSON in contract tests and diagnostic tooling;
- wire DTOs do not publish `schema_version` or `contract_version` fields;
- a tool that replaced a legacy raw-JSON method sets
  `Tool.TolerateUnknownArgs`, because the legacy params schemas left
  `additionalProperties` unset and therefore ignored unknown arguments;
- output fields are required unless explicitly modeled as optional;
- tools are registered in deterministic order;
- business failures become MCP tool errors, not private RPC envelopes;
- a migrated tool is removed from the legacy raw-JSON registry in the same
  change.

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
- Pydantic wire models do not publish `schema_version` or `contract_version`;
- `unique_list[T]` restores `uniqueItems: true`, which Pydantic does not project
  on its own;
- `$defs` are inlined, so a published schema stays a flat document;
- cohesive variants use typed union branches; remaining cross-field rules that
  Pydantic cannot project live in `@model_validator(mode="after")`.

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

## Counsel contract

All 10 counsel tools (`bazi_yongshen`, `natal_query`, `natal_liunian_query`,
`naming_surname`, `naming_pick`, `naming_characteracter`, `naming_compose`,
`naming_check`, `liuyao_ask`, `qimen_ask`) are typed contracts.
They are published together on the aggregate `/mcp` surface.

The yongshen decision is a standalone typed business capability. It returns
`selected_yongshen`, resolution `status`, conflicts, `rule_id`, policy provenance,
the `ge_ju_context`, branch `evidence`, and the two SHA-256 digests. The evidence
includes the 从格 effective/rooted counts used to verify or reject the candidate,
so a `needs_review` result is auditable over MCP. Naming reads the decision's
`selected_yongshen` plan and fails closed for `needs_review / unresolved`
results.

Natal factor evaluation is internal. `natal_query` and `natal_liunian_query`
accept the resolved subject and technique, build the chart and factors inside
the service, and return assertions / periods with evidence. They do not expose
the factor table, factor digest, or internal provenance artifact.

`counsel/tests/test_tool_contracts.py` inspects the current Pydantic schemas and
validates representative accepted and rejected values.
