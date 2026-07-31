package agent

import (
	"encoding/json"
	"strings"
)

// ── OpenRPC document generation ──────────────────────────────

type openRPCDoc struct {
	OpenRPC string        `json:"openrpc"`
	Info    openRPCInfo   `json:"info"`
	Methods []openRPCMeth `json:"methods"`
}

type openRPCInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

type openRPCMeth struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Params      json.RawMessage `json:"params"`
	Result      json.RawMessage `json:"result,omitempty"`
}

// OpenRPCDocument returns the full OpenRPC 1.4.1 document as raw JSON.
func (r *RPCRegistry) OpenRPCDocument() json.RawMessage {
	r.openrpcOnce.Do(func() {
		methods := make([]openRPCMeth, 0, len(r.names)+1)

		methods = append(methods, openRPCMeth{
			Name:        "rpc.discover",
			Description: "返回此 OpenRPC 1.4.1 document，包含所有可用 method 及参数定义。",
			Params:      json.RawMessage(`{"type":"object","properties":{},"description":"无需参数"}`),
			Result:      json.RawMessage(`{"type":"object","description":"OpenRPC 1.4.1 完整文档"}`),
		})

		for _, name := range r.names {
			m := r.methods[name]
			methods = append(methods, openRPCMeth{
				Name:        m.Name,
				Description: m.Description,
				Params:      m.Params,
				Result:      m.Result,
			})
		}

		doc := openRPCDoc{
			OpenRPC: "1.4.1",
			Info: openRPCInfo{
				Title:       "liki.hk JSON-RPC API",
				Version:     r.version,
				Description: "liki.hk Metaphysics Engine — 38 命理 API，让 AI agent 直接调用",
			},
			Methods: methods,
		}

		b, err := json.Marshal(doc)
		if err != nil {
			panic("marshal OpenRPC document: " + err.Error())
		}
		r.openrpcDoc = json.RawMessage(b)
	})
	return r.openrpcDoc
}

// DiscoverMethods returns a filtered OpenRPC document containing only the
// methods matching the given domain/name patterns (e.g. "bazi", "bazi.chart").
// A pattern matches the method itself and all methods under its dot-prefix.
// Empty patterns return the full document.
func (r *RPCRegistry) DiscoverMethods(patterns []string) json.RawMessage {
	full := r.OpenRPCDocument()
	var doc openRPCDoc
	if err := json.Unmarshal(full, &doc); err != nil {
		return full
	}

	if len(patterns) == 0 {
		return full
	}

	wanted := make(map[string]bool)
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		p = strings.TrimSuffix(p, ".")
		if p == "" {
			continue
		}
		prefix := p + "."
		for _, m := range doc.Methods {
			if m.Name == p || strings.HasPrefix(m.Name, prefix) {
				wanted[m.Name] = true
			}
		}
	}

	filtered := make([]openRPCMeth, 0, len(wanted))
	for _, m := range doc.Methods {
		if wanted[m.Name] {
			filtered = append(filtered, m)
		}
	}

	out := openRPCDoc{
		OpenRPC: doc.OpenRPC,
		Info:    doc.Info,
		Methods: filtered,
	}
	b, err := json.Marshal(out)
	if err != nil {
		return full
	}
	return json.RawMessage(b)
}
