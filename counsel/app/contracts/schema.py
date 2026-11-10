"""Contract plumbing for the counsel MCP tool surface.

The Pydantic v2 models in this package are the single source of truth for every
counsel tool: the published MCP `inputSchema` / `outputSchema`, the runtime
argument and result validation, and the JSON serialization all project from the
same models. The published schemas are emitted by Pydantic's own JSON Schema
generator; nothing about a tool's shape is written by hand.
"""

from __future__ import annotations

import hashlib
import json
from dataclasses import dataclass, field
from typing import Annotated, Any, TypeVar

from pydantic import AfterValidator, TypeAdapter
from pydantic.json_schema import GenerateJsonSchema

T = TypeVar("T")

#: JSON Schema keyword Pydantic does not project on its own.
_UNIQUE_ITEMS = "uniqueItems"
_UNSET = object()


class UniqueItems:
    """Annotate a list field with JSON Schema ``uniqueItems: true``."""

    def __get_pydantic_json_schema__(self, schema: Any, handler: Any) -> Any:
        schema = handler(schema)
        schema[_UNIQUE_ITEMS] = True
        return schema


def _reject_duplicates(values: list[Any]) -> list[Any]:
    seen: set[str] = set()
    for value in values:
        key = json.dumps(value, sort_keys=True, ensure_ascii=False, default=str)
        if key in seen:
            raise ValueError("duplicate items are not allowed")
        seen.add(key)
    return values


#: ``unique_list[T]`` = ``list[T]`` with ``uniqueItems: true`` semantics.
unique_list = Annotated[list[T], UniqueItems(), AfterValidator(_reject_duplicates)]


def _inline_refs(node: Any, defs: dict[str, Any], active: frozenset[str] = frozenset()) -> Any:
    """Inline ``$defs`` references so the published schema stays a flat document."""
    if isinstance(node, dict):
        ref = node.get("$ref")
        if isinstance(ref, str) and ref.startswith("#/$defs/"):
            name = ref.rsplit("/", 1)[-1]
            if name in active:
                raise ValueError(f"published schemas must be acyclic; {name!r} is recursive")
            target = json.loads(json.dumps(defs[name]))
            target.update({key: value for key, value in node.items() if key != "$ref"})
            return _inline_refs(target, defs, active | {name})
        return {key: _inline_refs(value, defs, active) for key, value in node.items()}
    if isinstance(node, list):
        return [_inline_refs(value, defs, active) for value in node]
    return node


def _normalize_nullable_any_of(node: Any) -> None:
    """Rewrite ``[null, T]`` unions to the equivalent compact ``type`` array."""
    if isinstance(node, dict):
        branches = node.get("anyOf")
        if isinstance(branches, list) and len(branches) == 2:
            non_null = None
            null_count = 0
            for branch in branches:
                if not isinstance(branch, dict):
                    continue
                if set(branch) == {"type"} and branch["type"] == "null":
                    null_count += 1
                    continue
                non_null = branch
            if null_count == 1 and isinstance(non_null, dict):
                kind = non_null.get("type")
                if isinstance(kind, str) and kind != "null":
                    replacement = {
                        key: value for key, value in non_null.items() if key != "type"
                    }
                    replacement["type"] = ["null", kind]
                    replacement.update(
                        {
                            key: value
                            for key, value in node.items()
                            if key not in {"anyOf", "type"}
                        }
                    )
                    node.clear()
                    node.update(replacement)
        for value in node.values():
            _normalize_nullable_any_of(value)
    elif isinstance(node, list):
        for value in node:
            _normalize_nullable_any_of(value)


def _schema_key(schema: dict[str, Any]) -> str:
    canonical = json.dumps(schema, sort_keys=True, ensure_ascii=False, separators=(",", ":"))
    return hashlib.sha256(canonical.encode()).hexdigest()[:12]


def _enum_schema_key(schema: dict[str, Any]) -> str:
    reader_metadata = {"description", "title", "default", "examples"}
    comparable = {
        key: value for key, value in schema.items() if key not in reader_metadata
    }
    values = comparable.get("enum")
    if isinstance(values, list) and all(isinstance(value, str) for value in values):
        comparable["enum"] = sorted(values)
    return _schema_key(comparable)


def _collect_enum_schemas(node: Any, counts: dict[str, int]) -> None:
    if isinstance(node, dict):
        if "enum" in node:
            key = _enum_schema_key(node)
            counts[key] = counts.get(key, 0) + 1
        for value in node.values():
            _collect_enum_schemas(value, counts)
    elif isinstance(node, list):
        for value in node:
            _collect_enum_schemas(value, counts)


def _hoist_enum_schemas(
    node: Any,
    definitions: dict[str, Any],
    counts: dict[str, int],
) -> Any:
    """Replace exact repeated enum schemas with shared, resolvable ``$defs``."""
    if isinstance(node, dict):
        if "enum" in node and counts[_enum_schema_key(node)] >= 2:
            name = f"Enum_{_enum_schema_key(node)}"
            if name not in definitions:
                definition = {
                    key: value for key, value in node.items() if key != "description"
                }
                if isinstance(definition.get("enum"), list) and all(
                    isinstance(value, str) for value in definition["enum"]
                ):
                    definition["enum"] = sorted(definition["enum"])
                definitions[name] = definition
            ref: dict[str, Any] = {"$ref": f"#/$defs/{name}"}
            # Keep reader-facing metadata at the field itself so catalogs do not
            # have to dereference a ref merely to show why the field exists.
            for key in ("description", "title", "default", "examples"):
                if key in node:
                    ref[key] = node[key]
            return ref
        return {
            key: _hoist_enum_schemas(value, definitions, counts)
            for key, value in node.items()
        }
    if isinstance(node, list):
        return [_hoist_enum_schemas(value, definitions, counts) for value in node]
    return node


def _slim_published_schema(schema: dict[str, Any]) -> dict[str, Any]:
    """Apply semantics-preserving publication compaction to a flat schema.

    Runtime validation is owned by the Pydantic models; this only compactly
    represents the same JSON Schema. Repeated enum schemas are shared through
    ``$defs`` while field descriptions remain at the field site.
    """
    _normalize_nullable_any_of(schema)
    counts: dict[str, int] = {}
    _collect_enum_schemas(schema, counts)
    existing = schema.pop("$defs", {})
    definitions = dict(existing)
    slim = _hoist_enum_schemas(schema, definitions, counts)
    if definitions:
        slim["$defs"] = definitions
    return slim


class ContractSchema(GenerateJsonSchema):
    """Pydantic's generator, restricted to what a published tool schema needs.

    Field titles and derived model titles are implementation noise for a tool
    schema, so only a model that declares ``ConfigDict(title=...)`` keeps one, and
    ``default: null`` is dropped because an optional contract field means
    "absent" rather than "send null".
    """

    def field_title_should_be_set(self, schema: Any) -> bool:
        return False

    def model_schema(self, schema: Any, **kwargs: Any) -> Any:  # type: ignore[override]
        json_schema = super().model_schema(schema, **kwargs)
        model = schema.get("cls")
        explicit = (getattr(model, "model_config", None) or {}).get("title")
        if json_schema.get("title") != explicit:
            json_schema.pop("title", None)
        return json_schema

    def default_schema(self, schema: Any) -> Any:
        json_schema = super().default_schema(schema)
        if json_schema.get("default", _UNSET) is None:
            json_schema.pop("default")
        return json_schema


def emit_schema(annotation: Any, *, description: str | None = None) -> dict[str, Any]:
    """Project a Pydantic annotation to its published JSON Schema.

    ``description`` exists because Pydantic v2 cannot attach a description to a
    model itself, and a tool output schema has no field to hang one on; the
    generator is the only place that can add it after the fact.
    """
    adapter = TypeAdapter(annotation)
    schema = adapter.json_schema(schema_generator=ContractSchema)
    defs = schema.pop("$defs", {})
    schema = _inline_refs(schema, defs)
    # MCP tool input schemas must be typed objects. A union whose branches are
    # all objects projects as bare anyOf; decorate it with the common object
    # type while preserving branch-specific properties and required sets.
    if "type" not in schema and schema.get("anyOf") and all(
        isinstance(branch, dict) and branch.get("type") == "object"
        for branch in schema["anyOf"]
    ):
        schema["type"] = "object"
        if all(branch.get("additionalProperties") is False for branch in schema["anyOf"]):
            schema["additionalProperties"] = False
    schema = _slim_published_schema(schema)
    if description is not None and "description" not in schema:
        schema["description"] = description
    return schema


@dataclass
class ToolContract:
    """One MCP tool: description plus the models that own its wire contract.

    Most counsel tools are deterministic, side-effect-free computations over
    their arguments, so ``annotations`` defaults to readOnly + idempotent.
    Event-creating tools (notably auto Liuyao casting) override idempotent to
    false at the contract that owns the semantics.
    """

    name: str
    description: str
    input: Any
    output: Any = None
    output_description: str | None = None
    annotations: dict[str, Any] = field(default_factory=lambda: {
        "readOnlyHint": True,
        "idempotentHint": True,
    })
    _input_schema: dict[str, Any] | None = field(default=None, init=False, repr=False)
    _output_schema: dict[str, Any] | None = field(default=None, init=False, repr=False)

    @property
    def input_schema(self) -> dict[str, Any]:
        if self._input_schema is None:
            self._input_schema = emit_schema(self.input)
        return self._input_schema

    @property
    def output_schema(self) -> dict[str, Any] | None:
        if self.output is None:
            return None
        if self._output_schema is None:
            self._output_schema = emit_schema(self.output, description=self.output_description)
        return self._output_schema
