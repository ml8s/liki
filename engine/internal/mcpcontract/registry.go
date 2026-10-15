// Package mcpcontract owns the MCP-native typed tool registry.
//
// The registry is intentionally small: explicit wire DTOs define the public
// schemas, and the same DTOs validate runtime input and output. It is not a
// general schema framework and does not expose legacy RPC envelopes.
package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
)

// Tool describes a typed MCP tool. Name uses the legacy registry convention
// when replacing an existing tool (for example "time.now"); the MCP surface
// converts it to a valid MCP name.
type Tool[In, Out any] struct {
	Name        string
	Description string
	Handler     func(context.Context, In) (Out, error)
	// TolerateUnknownArgs relaxes input validation for tools that replaced a
	// legacy raw-JSON method. The legacy registry validated params against the
	// raw JSON Schemas, which left `additionalProperties` unset and therefore
	// ignored unknown arguments: counsel passes tianwen.time's longitude to
	// bazi.chart and chains a ziwei.fullchart payload into ziwei.daxian. Tools
	// that were written directly against this typed registry keep rejecting
	// unknown arguments.
	TolerateUnknownArgs bool
	// KeepClosedInputPaths keeps individual input objects closed even when
	// TolerateUnknownArgs is set, because the legacy params schema declared
	// `additionalProperties: false` there (qimen.chart's root, and the flow
	// tools' target_lunar). Paths are JSON Pointers into the input schema:
	// "" is the root object.
	KeepClosedInputPaths []string
}

type registration struct {
	name         string
	description  string
	inputSchema  json.RawMessage
	outputSchema json.RawMessage
	call         func(context.Context, json.RawMessage) (json.RawMessage, error)
}

// Registry holds typed MCP tools in registration order.
type Registry struct {
	tools map[string]*registration
	names []string
}

// New creates an empty typed registry.
func New() *Registry {
	return &Registry{tools: make(map[string]*registration)}
}

// register derives schemas from In and Out, then validates both DTOs at the
// MCP boundary.
func register[In, Out any](r *Registry, tool Tool[In, Out]) error {
	if tool.Name == "" {
		return fmt.Errorf("mcp tool name is empty")
	}
	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("duplicate mcp tool: %s", tool.Name)
	}
	if tool.Handler == nil {
		return fmt.Errorf("mcp tool %s: handler is nil", tool.Name)
	}

	inputResolved, inputSchema, err := inputSchemaFor[In](tool.TolerateUnknownArgs, tool.KeepClosedInputPaths)
	if err != nil {
		return fmt.Errorf("mcp tool %s: input schema: %w", tool.Name, err)
	}
	outputSchemaDocument, outputResolved, outputSchema, err := schemaFor[Out]()
	if err != nil {
		return fmt.Errorf("mcp tool %s: output schema: %w", tool.Name, err)
	}

	// MCP conformance in this service exposes object roots only. Array and
	// scalar outputs keep runtime DTO validation but preserve the existing
	// wire shape without a declared outputSchema.
	exposedOutputSchema := json.RawMessage(nil)
	if outputSchemaDocument.Type == "object" {
		exposedOutputSchema = outputSchema
	}

	r.tools[tool.Name] = &registration{
		name:         tool.Name,
		description:  tool.Description,
		inputSchema:  inputSchema,
		outputSchema: exposedOutputSchema,
		call: func(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
			input, err := decodeInput[In](inputResolved, raw)
			if err != nil {
				return nil, fmt.Errorf("invalid arguments for %s: %w", tool.Name, err)
			}
			output, err := tool.Handler(ctx, input)
			if err != nil {
				return nil, err
			}
			return encodeOutput(outputResolved, output)
		},
	}
	r.names = append(r.names, tool.Name)
	return nil
}

// Names returns tool names in deterministic registration order.
func (r *Registry) Names() []string {
	return append([]string(nil), r.names...)
}

// Definition exposes metadata and generated schemas for an MCP server.
type Definition struct {
	Name         string
	Description  string
	InputSchema  json.RawMessage
	OutputSchema json.RawMessage
}

// Definition returns a tool's public MCP contract.
func (r *Registry) Definition(name string) (Definition, bool) {
	tool, ok := r.tools[name]
	if !ok {
		return Definition{}, false
	}
	return Definition{
		Name:         tool.name,
		Description:  tool.description,
		InputSchema:  tool.inputSchema,
		OutputSchema: tool.outputSchema,
	}, true
}

// Call validates input, invokes the typed handler, validates output, and
// returns the output DTO as JSON.
func (r *Registry) Call(ctx context.Context, name string, arguments json.RawMessage) (json.RawMessage, error) {
	tool, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return tool.call(ctx, arguments)
}

// inputSchemaFor derives a tool's documented input schema and the schema used
// to validate incoming arguments.
//
// The documented schema always keeps the typed, closed shape of the DTO. When
// tolerateUnknownArgs is set - for tools that replaced a legacy raw-JSON method
// - validation additionally ignores unknown properties, because the legacy
// schemas left `additionalProperties` unset (see Tool.TolerateUnknownArgs).
func inputSchemaFor[T any](tolerateUnknownArgs bool, keepClosed []string) (*jsonschema.Resolved, json.RawMessage, error) {
	schema, err := jsonschema.For[T](forOptions)
	if err != nil {
		return nil, nil, err
	}
	documented, err := schema.MarshalJSON()
	if err != nil {
		return nil, nil, err
	}
	if tolerateUnknownArgs {
		openObjectProperties(schema, keepClosed)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		return nil, nil, err
	}
	return resolved, json.RawMessage(documented), nil
}

// openObjectProperties drops the `additionalProperties: false` constraint that
// jsonschema-go derives from Go structs, leaving objects open to extra
// properties. Objects at one of the keepClosed JSON Pointer paths, and
// struct-valued map entries, keep their schema: only the false schema (the "no
// extra property is valid" form) is removed.
func openObjectProperties(schema *jsonschema.Schema, keepClosed []string) {
	closed := make(map[string]bool, len(keepClosed))
	for _, path := range keepClosed {
		closed[path] = true
	}
	seen := map[*jsonschema.Schema]bool{}
	var walk func(*jsonschema.Schema, string)
	walk = func(node *jsonschema.Schema, path string) {
		if node == nil || seen[node] {
			return
		}
		seen[node] = true
		if !closed[path] {
			if isFalseSchema(node.AdditionalProperties) {
				node.AdditionalProperties = nil
			}
			if isFalseSchema(node.UnevaluatedProperties) {
				node.UnevaluatedProperties = nil
			}
		}
		for name, sub := range node.Properties {
			walk(sub, path+"/properties/"+name)
		}
		for name, sub := range node.PatternProperties {
			walk(sub, path+"/patternProperties/"+name)
		}
		for name, sub := range node.Defs {
			walk(sub, path+"/$defs/"+name)
		}
		for name, sub := range node.Definitions {
			walk(sub, path+"/definitions/"+name)
		}
		for index, sub := range node.PrefixItems {
			walk(sub, fmt.Sprintf("%s/prefixItems/%d", path, index))
		}
		for index, sub := range node.ItemsArray {
			walk(sub, fmt.Sprintf("%s/items/%d", path, index))
		}
		for index, sub := range node.AllOf {
			walk(sub, fmt.Sprintf("%s/allOf/%d", path, index))
		}
		for index, sub := range node.AnyOf {
			walk(sub, fmt.Sprintf("%s/anyOf/%d", path, index))
		}
		for index, sub := range node.OneOf {
			walk(sub, fmt.Sprintf("%s/oneOf/%d", path, index))
		}
		walk(node.Items, path+"/items")
		walk(node.AdditionalItems, path+"/additionalItems")
		walk(node.Contains, path+"/contains")
		walk(node.PropertyNames, path+"/propertyNames")
		walk(node.Not, path+"/not")
	}
	walk(schema, "")
}

// isFalseSchema reports whether the schema is the JSON Schema `false`
// subschema (the form jsonschema-go uses to close an object).
func isFalseSchema(schema *jsonschema.Schema) bool {
	return schema != nil && schema.Not != nil && reflect.DeepEqual(*schema.Not, jsonschema.Schema{})
}

func schemaFor[T any]() (*jsonschema.Schema, *jsonschema.Resolved, json.RawMessage, error) {
	schema, err := jsonschema.For[T](forOptions)
	if err != nil {
		return nil, nil, nil, err
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		return nil, nil, nil, err
	}
	raw, err := schema.MarshalJSON()
	if err != nil {
		return nil, nil, nil, err
	}
	return schema, resolved, json.RawMessage(raw), nil
}

func decodeInput[In any](resolved *jsonschema.Resolved, raw json.RawMessage) (In, error) {
	var input In
	if len(raw) == 0 || string(raw) == "null" {
		raw = json.RawMessage(`{}`)
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return input, err
	}
	if err := resolved.Validate(document); err != nil {
		return input, err
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return input, err
	}
	return input, nil
}

func encodeOutput[Out any](resolved *jsonschema.Resolved, output Out) (json.RawMessage, error) {
	raw, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	if err := resolved.Validate(document); err != nil {
		return nil, fmt.Errorf("output contract violation: %w", err)
	}
	return raw, nil
}
