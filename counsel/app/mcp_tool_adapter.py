"""Adapt the Pydantic tool contracts to the Python MCP SDK.

The models in :mod:`app.contracts` are the single source of truth: the published
``inputSchema``/``outputSchema``, the boundary argument validation and the result
validation all come from them. No JSON Schema is written or parsed by hand here.
"""

from __future__ import annotations

import copy
import inspect
import json
import time
import types
from typing import Any

import anyio
from mcp.server import MCPServer
from mcp.server.mcpserver.exceptions import ToolError
from mcp.types import CallToolResult, TextContent, ToolAnnotations
from pydantic import BaseModel, ConfigDict, TypeAdapter, ValidationError, create_model

from app.engine_mcp_client import MCPError as EngineMCPError
from app.metrics import observe_tool_request
from app.natal.tools.errors import LikiToolError

#: Sentinel for "argument not supplied", so absent optionals stay absent instead
#: of being sent as ``None`` (the tool implementations distinguish the two).
_UNSET = object()


def _metric_domain(tool: str) -> str:
    if tool.startswith(("bazi_", "ziwei_", "liuyao_", "qimen_", "naming_")):
        return tool.split("_", 1)[0]
    if tool in {"compute_factors", "natal_query", "period_query"}:
        return "natal"
    return "aggregate"


def _metric_status(error: BaseException | None) -> str:
    if error is None:
        return "success"
    if isinstance(error, ToolError):
        message = str(error)
        if message.startswith("invalid_arguments:"):
            return "invalid_arguments"
        if message.startswith("invalid_result:"):
            return "invalid_result"
        if message.startswith("engine_dependency_failed:"):
            return "engine_dependency_failed"
        return "tool_failed"
    return "unhandled_exception"


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
        for name, info in _input_model_fields(contract).items()
    }
    return create_model(
        f"{contract.name}_Arguments",
        __base__=RawToolArguments,
        **fields,
    )


def validate_arguments(contract: Any, arguments: dict) -> None:
    """Validate tool arguments at the boundary with the contract's Pydantic model."""
    try:
        TypeAdapter(contract.input).validate_python(arguments)
    except ValidationError as exc:
        details = []
        for error in exc.errors():
            path = ".".join(str(part) for part in error["loc"]) or "$"
            details.append(f"{path}: {error['msg']}")
        raise ToolError("invalid_arguments: " + "; ".join(details)) from exc


def _input_model_fields(contract: Any) -> dict[str, Any]:
    """Declared input fields; for a union input, the union of branch fields.

    The carrier only transports values — the union itself decides which branch
    (and therefore which required fields) applies during boundary validation.
    """
    annotation = contract.input
    if isinstance(annotation, type) and issubclass(annotation, BaseModel):
        return dict(annotation.model_fields)
    if isinstance(annotation, types.UnionType):
        fields: dict[str, Any] = {}
        for member in annotation.__args__:
            if isinstance(member, type) and issubclass(member, BaseModel):
                for name, info in member.model_fields.items():
                    fields.setdefault(name, info)
        return fields
    return {}


def tool_signature(contract: Any) -> inspect.Signature:
    """Synthesize an explicit signature for the generic MCP handler."""
    parameters: list[inspect.Parameter] = []
    for name, _info in _input_model_fields(contract).items():
        if not name.isidentifier() or inspect.iskeyword(name):
            raise ValueError(f"MCP tool parameter is not a Python identifier: {name!r}")
        parameters.append(
            inspect.Parameter(
                name,
                inspect.Parameter.POSITIONAL_OR_KEYWORD,
                annotation=Any,
                default=None,
            )
        )
    return inspect.Signature(parameters, return_annotation=Any)


def add_contract_tool(server: MCPServer, contract: Any, invoke) -> None:
    """Register one contract tool: Pydantic in, Pydantic out.

    ``invoke`` 是类型化纯函数，签名为 ``(**kwargs) -> Any``：边界先做入参
    校验（contract.input），以关键字实参直通调用（不做二次 dict 拆包），
    再做输出校验（contract.output，对齐 engine registry 的 encodeOutput）
    后才作为工具结果发布。
    """

    async def handler(**actual):
        arguments = {key: value for key, value in actual.items() if value is not _UNSET}
        started = time.perf_counter()
        terminal_error: BaseException | None = None
        # The model is the gate, not a coercion step: the raw arguments are what
        # the tool implementations receive, exactly as the JSON-schema boundary
        # behaved before.
        try:
            validate_arguments(contract, arguments)
            output = TypeAdapter(contract.output) if contract.output is not None else None
            try:
                result = await anyio.to_thread.run_sync(lambda: invoke(**arguments))
            except (ValueError, LikiToolError) as exc:
                raise ToolError(f"tool_failed: {exc}") from exc
            except EngineMCPError as exc:
                raise ToolError(f"engine_dependency_failed: {exc}") from exc
            if output is not None:
                try:
                    output.validate_python(result)
                except ValidationError as exc:
                    details = "; ".join(
                        ".".join(str(part) for part in error["loc"]) or "$"
                        for error in exc.errors()
                    )
                    raise ToolError(f"invalid_result: {details}") from exc
            # Keep the existing JSON-string wire shape even for list results. SDK list
            # conversion would otherwise expand the list into one text block per item.
            return CallToolResult(
                content=[
                    TextContent(type="text", text=json.dumps(result, ensure_ascii=False))
                ],
                structured_content=result,
            )
        except Exception as exc:
            terminal_error = exc
            raise
        finally:
            observe_tool_request(
                contract.name,
                _metric_domain(contract.name),
                _metric_status(terminal_error),
                time.perf_counter() - started,
            )

    handler.__signature__ = tool_signature(contract)

    server.add_tool(
        handler,
        name=contract.name,
        description=contract.description,
        annotations=ToolAnnotations(**contract.annotations),
    )
    registered = server._tool_manager.get_tool(contract.name)  # noqa: SLF001 - SDK has no schema setter
    if registered is None:
        raise RuntimeError(f"MCP tool was not registered: {contract.name}")
    registered.fn_metadata.arg_model = tool_argument_model(contract)
    if contract.output is not None:
        registered.fn_metadata.output_model = contract.output
        registered.fn_metadata.output_schema = copy.deepcopy(contract.output_schema)
    registered.parameters = copy.deepcopy(contract.input_schema)
