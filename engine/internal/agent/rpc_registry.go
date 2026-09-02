package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// RPC method schemas are defined as inline Go strings. They drive the OpenRPC
// document and parameter validation while remaining adjacent to their handlers.

// Common JSON Schema fragments (inline, no $ref — self-contained for AI agents).
const (
	schemaSolarTime = `{"type":"string","description":"真太阳时（RFC3339），从 tianwen.time 返回的 solar 字段","examples":["2026-07-20T12:00:00+08:00"]}`
	schemaGender    = `{"type":"string","enum":["male","female"],"description":"性别","examples":["male"]}`
)

func mustSchema(s string) json.RawMessage {
	if !json.Valid([]byte(s)) {
		panic("invalid JSON schema")
	}
	return json.RawMessage(s)
}

// envelopeSchema wraps a data schema in the standard {"_product":"...","data":<schema>} envelope.
func envelopeSchema(dataSchema string) json.RawMessage {
	return mustSchema(`{"type":"object","properties":{"_product":{"type":"string"},"data":` + dataSchema + `},"required":["_product","data"]}`)
}

// RPCMethod describes a single JSON-RPC method.
type RPCMethod struct {
	Name        string
	Description string
	Params      json.RawMessage // JSON Schema for params
	Result      json.RawMessage // JSON Schema for result (optional)
	Handler     func(context.Context, json.RawMessage) (json.RawMessage, error)

	paramsSchema *jsonschema.Schema // compiled params schema for validation
}

// RPCRegistry holds all registered JSON-RPC methods.
type RPCRegistry struct {
	methods map[string]*RPCMethod
	names   []string // registration order preserved for deterministic output

	version     string
	openrpcOnce sync.Once
	openrpcDoc  json.RawMessage
}

// SetVersion sets the engine version reported in the OpenRPC document.
// Called once at startup from the main package (BuildTime / VERSION file).
func (r *RPCRegistry) SetVersion(v string) {
	r.version = v
	r.openrpcOnce = sync.Once{} // reset cache so the doc is rebuilt with the new version
}

// NewRPCRegistry creates a registry with all external compute methods
// (count is locked by TestNewRPCRegistry_ExpectedMethodCount).
func NewRPCRegistry() *RPCRegistry {
	r := &RPCRegistry{methods: make(map[string]*RPCMethod, 32)}
	for _, m := range baziMethods {
		r.mustRegister(m)
	}
	for _, m := range ziweiMethods {
		r.mustRegister(m)
	}
	for _, m := range qimingMethods {
		r.mustRegister(m)
	}
	for _, m := range otherMethods {
		r.mustRegister(m)
	}
	return r
}

func (r *RPCRegistry) mustRegister(m RPCMethod) {
	if _, exists := r.methods[m.Name]; exists {
		panic("duplicate RPC method: " + m.Name)
	}
	if len(m.Params) > 0 {
		m.paramsSchema = compileParamsSchema(m.Name, m.Params)
	}
	r.methods[m.Name] = &m
	r.names = append(r.names, m.Name)
}

func compileParamsSchema(name string, raw json.RawMessage) *jsonschema.Schema {
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		log.Printf("rpc: %s: failed to parse params schema: %v", name, err)
		return nil
	}
	c := jsonschema.NewCompiler()
	url := "https://liki.hk/schemas/" + name + "/params.json"
	if err := c.AddResource(url, doc); err != nil {
		log.Printf("rpc: %s: failed to add resource: %v", name, err)
		return nil
	}
	sch, err := c.Compile(url)
	if err != nil {
		log.Printf("rpc: %s: failed to compile params schema: %v", name, err)
		return nil
	}
	return sch
}

// Execute runs the handler for the given method name with raw JSON params.
func (r *RPCRegistry) Execute(ctx context.Context, method string, params json.RawMessage) (json.RawMessage, error) {
	m, ok := r.methods[method]
	if !ok {
		return nil, &RPCError{Code: -32601, Message: "Method not found: " + method}
	}

	if m.paramsSchema != nil && len(params) > 0 {
		var paramsDoc any
		if err := json.Unmarshal(params, &paramsDoc); err != nil {
			return nil, &RPCError{Code: -32600, Message: "Invalid JSON: " + err.Error()}
		}
		if err := m.paramsSchema.Validate(paramsDoc); err != nil {
			return nil, &RPCError{Code: -32602, Message: fmt.Sprintf("Invalid params for %s: %v", method, err)}
		}
	}

	result, err := m.Handler(ctx, params)
	if err != nil {
		return nil, &RPCError{Code: -32000, Message: err.Error()}
	}
	return result, nil
}

// RPCError implements error and serializes as a JSON-RPC error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string { return e.Message }
