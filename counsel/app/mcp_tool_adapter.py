"""Adapt the Pydantic tool contracts to the Python MCP SDK.

The models in :mod:`app.contracts` are the single source of truth: the published
``inputSchema``/``outputSchema``, the boundary argument validation and the result
validation all come from them. No JSON Schema is written or parsed by hand here.
"""

from __future__ import annotations

import copy
import inspect
import json
from typing import Any

import anyio
from mcp.server import MCPServer
from mcp.server.mcpserver.exceptions import ToolError
from mcp.types import CallToolResult, TextContent
from pydantic import BaseModel, ConfigDict, ValidationError, create_model

from app.engine_mcp_client import MCPError as EngineMCPError
from app.natal.tools.errors import LikiToolError

#: Sentinel for "argument not supplied", so absent optionals stay absent instead
#: of being sent as ``None`` (the tool implementations distinguish the two).
_UNSET = object()


class RawToolArguments(BaseModel):
    """Carry the declared arguments verbatim; the contract model owns validation."""

    model_config = ConfigDict(extra="allow", arbitrary_types_allowed=True)

    def model_dump_one_level(self) -> dict[str, Any]:
        arguments = {
            field_info.alias or field_name: getattr(self, field_name)
            for field_name, field_info in type(self).model_fields.items()
        }
        arguments.update(self.__pydantic_extra__ or {})
        return arguments


def tool_argument_model(contract: Any) -> type[RawToolArguments]:
    """Build the SDK argument carrier from the contract's declared fields.

    The carrier only transports values; :func:`validate_arguments` runs the real
    validation through the contract model.
    """
    fields = {
        name: (Any, ... if info.is_required() else _UNSET)
        for name, info in contract.input.model_fields.items()
    }
    return create_model(
        f"{contract.name}_Arguments",
        __base__=RawToolArguments,
        **fields,
    )


def validate_arguments(contract: Any, arguments: dict) -> None:
    """Validate tool arguments at the boundary with the contract's Pydantic model."""
    try:
        contract.input.model_validate(arguments)
    except ValidationError as exc:
        details = []
        for error in exc.errors():
            path = ".".join(str(part) for part in error["loc"]) or "$"
            details.append(f"{path}: {error['msg']}")
        raise ToolError("invalid_arguments: " + "; ".join(details)) from exc


def tool_signature(contract: Any) -> inspect.Signature:
    """Synthesize an explicit signature for the generic MCP handler."""
    parameters: list[inspect.Parameter] = []
    for name, info in contract.input.model_fields.items():
        if not name.isidentifier() or inspect.iskeyword(name):
            raise ValueError(f"MCP tool parameter is not a Python identifier: {name!r}")
        parameters.append(
            inspect.Parameter(
                name,
                inspect.Parameter.POSITIONAL_OR_KEYWORD,
                annotation=Any,
                default=None if not info.is_required() else inspect.Parameter.empty,
            )
        )
    return inspect.Signature(parameters, return_annotation=Any)


def add_contract_tool(server: MCPServer, contract: Any, invoke) -> None:
    """Register one contract tool: Pydantic in, Pydantic out."""

    async def handler(**actual):
        arguments = {key: value for key, value in actual.items() if value is not _UNSET}
        # The model is the gate, not a coercion step: the raw arguments are what
        # the tool implementations receive, exactly as the JSON-schema boundary
        # behaved before.
        validate_arguments(contract, arguments)
        try:
            result = await anyio.to_thread.run_sync(invoke, arguments)
        except (ValueError, LikiToolError) as exc:
            raise ToolError(f"tool_failed: {exc}") from exc
        except EngineMCPError as exc:
            raise ToolError(f"engine_dependency_failed: {exc}") from exc
        # Keep the existing JSON-string wire shape even for list results. SDK list
        # conversion would otherwise expand the list into one text block per item.
        return CallToolResult(
            content=[TextContent(type="text", text=json.dumps(result, ensure_ascii=False))],
            structured_content=result,
        )

    handler.__signature__ = tool_signature(contract)

    server.add_tool(handler, name=contract.name, description=contract.description)
    registered = server._tool_manager.get_tool(contract.name)  # noqa: SLF001 - SDK has no schema setter
    if registered is None:
        raise RuntimeError(f"MCP tool was not registered: {contract.name}")
    registered.fn_metadata.arg_model = tool_argument_model(contract)
    if contract.output is not None:
        registered.fn_metadata.output_model = contract.output
        registered.fn_metadata.output_schema = copy.deepcopy(contract.output_schema)
    registered.parameters = copy.deepcopy(contract.input_schema)
