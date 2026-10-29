// Package mcpcontract owns the MCP-native typed tool registry.
//
// The registry is intentionally small: explicit wire DTOs define the public
// schemas and own runtime input validation. Output schemas remain published
// contracts and can be checked explicitly in tests, debug tooling, and staging;
// a reflective output check is not repeated on every production call.
package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
)

// Tool describes a typed MCP tool. Name is the public MCP tool name
// (snake_case, e.g. "bazi_chart").
type Tool[In, Out any] struct {
	Name        string
	Description string
	Handler     func(context.Context, In) (Out, error)
	// InputSchema overrides the schema derived from In when the public input is
	// a union of forms (e.g. tianwen.resolve's five input variants). The derived
	// schema is still used for boundary validation.
	InputSchema json.RawMessage
	// OutputSchema overrides the schema derived from Out. Used by series
	// tools whose handler returns SeriesResult[Out] (an items array for every
	// count), because the items element schema is the published Out.
	OutputSchema json.RawMessage
}

// RawTool is a tool whose domain value already marshals to the published Out
// contract. In keeps runtime input validation; Out is a phantom type used to
// derive and validate the output schema without reconstructing a wire DTO.
// Raw handlers must be covered by wire-fidelity fixtures.
type RawTool[In, Out any] struct {
	Name        string
	Description string
	Handler     func(context.Context, In) (json.RawMessage, error)
	InputSchema json.RawMessage
}

type registration struct {
	name        string
	description string
	schemas     *lazyToolSchemas
	call        func(context.Context, json.RawMessage) (json.RawMessage, error)
}

type lazyOutputValidator struct {
	schema   *jsonschema.Schema
	once     sync.Once
	resolved *jsonschema.Resolved
	err      error
}

func (v *lazyOutputValidator) validate(raw json.RawMessage) error {
	if err := v.resolve(); err != nil {
		return err
	}
	return validateJSONDocument(v.resolved, raw)
}

func (v *lazyOutputValidator) resolve() error {
	v.once.Do(func() { v.resolved, v.err = v.schema.Resolve(nil) })
	if v.err != nil {
		return fmt.Errorf("resolve output schema: %w", v.err)
	}
	return nil
}

type derivedSchemas struct {
	inputTree       *jsonschema.Schema
	inputResolved   *jsonschema.Resolved
	inputSchema     json.RawMessage
	outputSchema    json.RawMessage
	outputValidator *lazyOutputValidator
	fastInput       bool
	fastClosed      map[*jsonschema.Schema]bool
}

type lazyToolSchemas struct {
	derive  func() (derivedSchemas, error)
	once    sync.Once
	schemas derivedSchemas
	err     error
}

func (s *lazyToolSchemas) get() (derivedSchemas, error) {
	s.once.Do(func() {
		s.schemas, s.err = s.derive()
	})
	return s.schemas, s.err
}

func deriveToolSchemas[In, Out any](publicInputSchema json.RawMessage, publicOutputSchema json.RawMessage) (derivedSchemas, error) {
	inputTree, inputResolved, inputSchema, err := inputSchemaTreeFor[In]()
	if err != nil {
		return derivedSchemas{}, err
	}
	if len(publicInputSchema) > 0 {
		inputSchema = publicInputSchema
	}
	var outputSchema json.RawMessage
	var outputValidator *lazyOutputValidator
	if len(publicOutputSchema) > 0 {
		outputSchema = publicOutputSchema
		outputValidator, err = lazyValidatorFromRaw(publicOutputSchema)
		if err != nil {
			return derivedSchemas{}, fmt.Errorf("derive output schema: %w", err)
		}
	} else {
		outputSchema, outputValidator, err = outputSchemaFor[Out]()
	}
	if err != nil {
		return derivedSchemas{}, err
	}
	fastInput := supportsFastInputSchema(inputTree)
	fastClosed := map[*jsonschema.Schema]bool{}
	if fastInput {
		collectClosedInputSchemas(inputTree, fastClosed)
	}
	return derivedSchemas{
		inputTree:       inputTree,
		inputResolved:   inputResolved,
		inputSchema:     inputSchema,
		outputSchema:    outputSchema,
		outputValidator: outputValidator,
		fastInput:       fastInput,
		fastClosed:      fastClosed,
	}, nil
}

// mustRegisterRaw registers a raw-output tool and panics on contract drift.
func mustRegisterRaw[In, Out any](registry *Registry, tool RawTool[In, Out]) {
	if err := registerRaw[In, Out](registry, tool); err != nil {
		panic(fmt.Sprintf("register engine MCP contract %s: %v", tool.Name, err))
	}
}

func registerRaw[In, Out any](r *Registry, tool RawTool[In, Out]) error {
	if tool.Name == "" {
		return fmt.Errorf("mcp tool name is empty")
	}
	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("duplicate mcp tool: %s", tool.Name)
	}
	if tool.Handler == nil {
		return fmt.Errorf("mcp tool %s: handler is nil", tool.Name)
	}

	schemas := &lazyToolSchemas{
		derive: func() (derivedSchemas, error) {
			return deriveToolSchemas[In, Out](tool.InputSchema, nil)
		},
	}

	r.tools[tool.Name] = &registration{
		name:        tool.Name,
		description: tool.Description,
		schemas:     schemas,
		call: func(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
			derived, err := schemas.get()
			if err != nil {
				return nil, fmt.Errorf("invalid arguments for %s: derive schemas: %w", tool.Name, err)
			}
			input, err := decodeInputForSchema[In](derived, raw)
			if err != nil {
				return nil, fmt.Errorf("invalid arguments for %s: %w", tool.Name, err)
			}
			return tool.Handler(ctx, input)
		},
	}
	r.names = append(r.names, tool.Name)
	return nil
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

// register creates a lazy schema derivation for In and Out. Input is validated
// on every call; output validation is exposed as an explicit contract check.
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

	schemas := &lazyToolSchemas{
		derive: func() (derivedSchemas, error) {
			return deriveToolSchemas[In, Out](tool.InputSchema, tool.OutputSchema)
		},
	}

	// The current MCP SDK supports object, array, scalar, and composition
	// roots in outputSchema. Expose every root so clients can bootstrap the
	// full result shape without a probe call or hard-coded documentation.
	// Schema bytes are materialized by Definition below.

	r.tools[tool.Name] = &registration{
		name:        tool.Name,
		description: tool.Description,
		schemas:     schemas,
		call: func(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
			derived, err := schemas.get()
			if err != nil {
				return nil, fmt.Errorf("invalid arguments for %s: derive schemas: %w", tool.Name, err)
			}
			input, err := decodeInputForSchema[In](derived, raw)
			if err != nil {
				return nil, fmt.Errorf("invalid arguments for %s: %w", tool.Name, err)
			}
			output, err := tool.Handler(ctx, input)
			if err != nil {
				return nil, err
			}
			return encodeOutput(output)
		},
	}
	r.names = append(r.names, tool.Name)
	return nil
}

// Names returns tool names in deterministic registration order.
func (r *Registry) Names() []string {
	return append([]string(nil), r.names...)
}

// MaterializeSchemas derives every tool schema concurrently and returns the
// first deterministic registration-order error. The process calls it before it
// starts serving so a broken contract fails fast instead of surfacing on the
// first request. The worker count follows GOMAXPROCS, which also respects a
// container CPU quota.
func (r *Registry) MaterializeSchemas() error {
	type work struct {
		index int
		name  string
	}
	input := make(chan work)
	errors := make([]error, len(r.names))
	workers := runtime.GOMAXPROCS(0)
	if workers > len(r.names) {
		workers = len(r.names)
	}
	if workers == 0 {
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(workers)
	for worker := 0; worker < workers; worker++ {
		go func() {
			defer wg.Done()
			for item := range input {
				tool, exists := r.tools[item.name]
				if !exists {
					errors[item.index] = fmt.Errorf("unknown tool: %s", item.name)
					continue
				}
				schemas, err := tool.schemas.get()
				if err != nil {
					errors[item.index] = err
					continue
				}
				if len(schemas.outputSchema) == 0 {
					errors[item.index] = fmt.Errorf("mcp tool %s: empty output schema", item.name)
					continue
				}
				if len(schemas.inputSchema) == 0 {
					errors[item.index] = fmt.Errorf("mcp tool %s: empty input schema", item.name)
					continue
				}
				errors[item.index] = schemas.outputValidator.resolve()
			}
		}()
	}
	for index, name := range r.names {
		input <- work{index: index, name: name}
	}
	close(input)
	wg.Wait()

	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

// Definition exposes metadata and generated schemas for an MCP server.
type Definition struct {
	Name         string
	Description  string
	InputSchema  json.RawMessage
	OutputSchema json.RawMessage
}

// Definition returns a tool's public MCP contract. The process materializes all
// definitions before serving; laziness keeps registration cheap for focused
// registry construction in tests and benchmarks.
func (r *Registry) Definition(name string) (Definition, error) {
	tool, ok := r.tools[name]
	if !ok {
		return Definition{}, fmt.Errorf("unknown tool: %s", name)
	}
	schemas, err := tool.schemas.get()
	if err != nil {
		return Definition{}, fmt.Errorf("mcp tool %s: derive schemas: %w", tool.name, err)
	}
	return Definition{
		Name:         tool.name,
		Description:  tool.description,
		InputSchema:  schemas.inputSchema,
		OutputSchema: schemas.outputSchema,
	}, nil
}

// Call validates untrusted input, invokes the typed handler, and marshals the
// output DTO. Use ValidateOutput in contract tests or diagnostic tooling to
// check the actual JSON document against the published output schema.
func (r *Registry) Call(ctx context.Context, name string, arguments json.RawMessage) (json.RawMessage, error) {
	tool, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return tool.call(ctx, arguments)
}

// ValidateOutput checks an actual JSON result against a tool's published
// output schema. It is intentionally separate from Call: output is produced by
// trusted typed code and is covered by contract tests, so production requests
// do not reconstruct the document for a reflective validation pass.
func (r *Registry) ValidateOutput(name string, output json.RawMessage) error {
	tool, ok := r.tools[name]
	if !ok {
		return fmt.Errorf("unknown tool: %s", name)
	}
	schemas, err := tool.schemas.get()
	if err != nil {
		return fmt.Errorf("mcp tool %s: derive schemas: %w", tool.name, err)
	}
	return schemas.outputValidator.validate(output)
}

// inputSchemaFor derives a tool's input schema: the typed, closed shape of the
// DTO, used both as the documented contract and for boundary validation.
func inputSchemaFor[T any]() (*jsonschema.Resolved, json.RawMessage, error) {
	_, resolved, raw, err := inputSchemaTreeFor[T]()
	return resolved, raw, err
}

func inputSchemaTreeFor[T any]() (*jsonschema.Schema, *jsonschema.Resolved, json.RawMessage, error) {
	schema, err := jsonschema.For[T](forOptions)
	if err != nil {
		return nil, nil, nil, err
	}
	documented, err := schema.MarshalJSON()
	if err != nil {
		return nil, nil, nil, err
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		return nil, nil, nil, err
	}
	return schema, resolved, json.RawMessage(documented), nil
}

func outputSchemaFor[T any]() (json.RawMessage, *lazyOutputValidator, error) {
	schema, err := jsonschema.For[T](forOptions)
	if err != nil {
		return nil, nil, err
	}
	raw, err := schema.MarshalJSON()
	if err != nil {
		return nil, nil, err
	}
	return json.RawMessage(raw), &lazyOutputValidator{schema: schema}, nil
}

// lazyValidatorFromRaw builds an output validator from a user-provided raw
// JSON schema (e.g. a oneOf union for series tools).
func lazyValidatorFromRaw(raw json.RawMessage) (*lazyOutputValidator, error) {
	var schema jsonschema.Schema
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, fmt.Errorf("unmarshal output schema: %w", err)
	}
	return &lazyOutputValidator{schema: &schema}, nil
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

func encodeOutput[Out any](output Out) (json.RawMessage, error) {
	raw, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func validateJSONDocument(resolved *jsonschema.Resolved, raw json.RawMessage) error {
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return err
	}
	if err := resolved.Validate(document); err != nil {
		return fmt.Errorf("output contract violation: %w", err)
	}
	return nil
}
