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
- output fields are required unless explicitly modeled as optional;
- tools are registered in deterministic order;
- business failures become MCP tool errors, not private RPC envelopes;
- a migrated tool is removed from the legacy raw-JSON registry in the same
  change.

## Migration

`time_now` is the reference implementation. Existing tools continue through
the compatibility registry until they are migrated one by one. A tool is
migrated only when its explicit wire DTOs cover the real MCP output and its
schema, description, validation, and `structuredContent` contract tests pass.

Python MCP services will use Pydantic v2 models for the same role: schema
generation, input validation, output validation, and JSON serialization.
