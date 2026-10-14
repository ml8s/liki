"""Adapt JSON-schema tool manifests to the Python MCP SDK."""

from __future__ import annotations

import copy
import inspect
import json
from typing import Any

import anyio
from jsonschema import validators
from mcp.server import MCPServer
from mcp.server.mcpserver.exceptions import ToolError
from mcp.types import TextContent
from pydantic import BaseModel, ConfigDict, create_model

from app.engine_mcp_client import MCPError as EngineMCPError
from app.natal.tools.errors import LikiToolError

_UNSET = object()


class RawToolArguments(BaseModel):
    """Preserve undeclared arguments for validation against the manifest."""

    model_config = ConfigDict(extra="allow", arbitrary_types_allowed=True)

    def model_dump_one_level(self) -> dict[str, Any]:
        arguments = {
            field_info.alias or field_name: getattr(self, field_name)
            for field_name, field_info in type(self).model_fields.items()
        }
        arguments.update(self.__pydantic_extra__ or {})
        return arguments


def raw_argument_model(tool_name: str, schema: dict) -> type[RawToolArguments]:
    properties = schema.get("properties", {})
    required = set(schema.get("required", []))
    fields = {}
    for name in properties:
        default = properties[name].get("default", _UNSET)
        fields[name] = (Any, ... if name in required else default)
    return create_model(
        f"{tool_name}_RawArguments",
        __base__=RawToolArguments,
        **fields,
    )


def validate_arguments(schema: dict, arguments: dict) -> None:
    """Validate tool arguments at the boundary using the manifest schema."""
    validator_class = validators.validator_for(schema)
    validator_class.check_schema(schema)
    validator = validator_class(schema)
    errors = sorted(
        validator.iter_errors(arguments),
        key=lambda error: (
            ".".join(str(part) for part in error.absolute_path),
            error.message,
        ),
    )
    if not errors:
        return
    details = []
    for error in errors:
        path = ".".join(str(part) for part in error.absolute_path) or "$"
        details.append(f"{path}: {error.message}")
    raise ToolError("invalid_arguments: " + "; ".join(details)) from errors[0]


def tool_signature(schema: dict) -> inspect.Signature:
    """Synthesize an explicit signature for the generic MCP handler."""
    required = set(schema.get("required", []))
    parameters: list[inspect.Parameter] = []
    for name in schema.get("properties", {}):
        if not name.isidentifier() or inspect.iskeyword(name):
            raise ValueError(f"MCP tool parameter is not a Python identifier: {name!r}")
        parameters.append(
            inspect.Parameter(
                name,
                inspect.Parameter.POSITIONAL_OR_KEYWORD,
                annotation=Any,
                default=None if name not in required else inspect.Parameter.empty,
            )
        )
    return inspect.Signature(parameters, return_annotation=str)


def add_schema_tool(
    server: MCPServer,
    name: str,
    description: str,
    schema: dict,
    invoke,
    *,
    output_schema: dict | None = None,
) -> None:
    """Register a tool while preserving its original JSON schema."""

    output_validator = validators.validator_for(output_schema) if output_schema else None
    if output_validator is not None:
        output_validator.check_schema(output_schema)
        output_validator = output_validator(output_schema)

    async def handler(**actual):
        arguments = {
            key: value for key, value in actual.items() if value is not _UNSET
        }
        if name == "naming_compose":
            arguments.setdefault("second", [])
        validate_arguments(schema, arguments)
        try:
            result = await anyio.to_thread.run_sync(invoke, arguments)
        except (ValueError, LikiToolError) as exc:
            raise ToolError(f"tool_failed: {exc}") from exc
        except EngineMCPError as exc:
            raise ToolError(f"engine_dependency_failed: {exc}") from exc
        if output_validator is not None:
            output_validator.validate(result)
        # Keep the existing JSON-string wire shape even for list results. SDK list
        # conversion would otherwise expand the list into one text block per item.
        from mcp.types import CallToolResult

        return CallToolResult(
            content=[TextContent(type="text", text=json.dumps(result, ensure_ascii=False))],
            structured_content=result,
        )

    handler.__signature__ = tool_signature(schema)

    server.add_tool(handler, name=name, description=description)
    registered = server._tool_manager.get_tool(name)  # noqa: SLF001 - SDK has no schema setter
    if registered is None:
        raise RuntimeError(f"MCP tool was not registered: {name}")
    registered.fn_metadata.arg_model = raw_argument_model(name, schema)
    if output_schema is not None:
        registered.fn_metadata.output_schema = copy.deepcopy(output_schema)
        registered.fn_metadata.output_model = Any
    registered.parameters = copy.deepcopy(schema)
