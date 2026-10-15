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

// forOptions adds reusable wire-level constraints that Go struct tags cannot
// express through this library. Field-level descriptions still come from the
// DTO's `jsonschema` tags.
var forOptions = &jsonschema.ForOptions{TypeSchemas: map[reflect.Type]*jsonschema.Schema{
	reflect.TypeFor[RFC3339Time](): {Type: "string", Format: "date-time"},
	reflect.TypeFor[DateString]():  {Type: "string", Format: "date"},
	reflect.TypeFor[CityName]():    {Type: "string", MinLength: jsonschema.Ptr(1), MaxLength: jsonschema.Ptr(128)},
	reflect.TypeFor[HuangliEvent](): {
		Type: "string",
		Enum: []any{
			"wedding", "engage", "opening", "sign", "move", "travel", "build",
			"exam", "medical", "sacrifice", "cleaning", "renovation",
			"bed_install", "income", "funeral",
		},
	},
	reflect.TypeFor[HuangliCount](): {Type: "integer", Minimum: jsonschema.Ptr(float64(1)), Maximum: jsonschema.Ptr(float64(30))},
}}

// Tool describes a typed MCP tool. Name uses the legacy registry convention
// when replacing an existing tool (for example "time.now"); the MCP surface
// converts it to a valid MCP name.
type Tool[In, Out any] struct {
	Name        string
	Description string
	Handler     func(context.Context, In) (Out, error)
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

	_, inputResolved, inputSchema, err := schemaFor[In]()
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

// Has reports whether a typed replacement exists for the named registry tool.
func (r *Registry) Has(name string) bool {
	_, ok := r.tools[name]
	return ok
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
