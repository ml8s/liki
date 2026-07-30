package http

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"liki-engine/internal/agent"
)

// RPCError is an alias for agent.RPCError for external use.
type RPCError = agent.RPCError

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      any             `json:"id"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  any             `json:"result,omitempty"`
	Error   *agent.RPCError `json:"error,omitempty"`
	ID      any             `json:"id"`
}

func HandleRPC(reg *agent.RPCRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method != http.MethodPost {
			writeRPC(w, rpcResponse{Error: &agent.RPCError{Code: -32600, Message: "only POST allowed"}, ID: nil})
			return
		}

		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeRPC(w, rpcResponse{Error: &agent.RPCError{Code: -32700, Message: "Parse error"}, ID: nil})
			return
		}

		if req.JSONRPC != "2.0" {
			writeRPC(w, rpcResponse{Error: &agent.RPCError{Code: -32600, Message: "jsonrpc must be \"2.0\""}, ID: req.ID})
			return
		}
		if req.Method == "" {
			writeRPC(w, rpcResponse{Error: &agent.RPCError{Code: -32600, Message: "method is required"}, ID: req.ID})
			return
		}
		if req.ID == nil {
			writeRPC(w, rpcResponse{Error: &agent.RPCError{Code: -32600, Message: "id is required (notifications not supported)"}, ID: nil})
			return
		}

		if len(req.Params) > 0 && req.Params[0] == '[' {
			writeRPC(w, rpcResponse{Error: &agent.RPCError{Code: -32600, Message: "positional params not supported, use an object"}, ID: req.ID})
			return
		}
		if len(req.Params) == 0 || bytes.Equal(req.Params, []byte("null")) {
			req.Params = json.RawMessage(`{}`)
		}

		// rpc.discover handled at HTTP layer, not in the registry
		if req.Method == "rpc.discover" {
			var fp struct{ Methods string `json:"methods,omitempty"` }
			json.Unmarshal(req.Params, &fp)

			// 从全量 doc 取出方法名列表
			doc := reg.OpenRPCDocument()
			var all struct {
				Methods []struct {
					Name string `json:"name"`
				} `json:"methods"`
			}
			json.Unmarshal(doc, &all)
			allNames := make([]string, len(all.Methods))
			for i, m := range all.Methods { allNames[i] = m.Name }

			if fp.Methods == "" {
				writeRPC(w, rpcResponse{Result: doc, ID: req.ID})
				return
			}

			// 不递归，一次 HasPrefix 匹配自身或下级
			patterns := strings.Split(fp.Methods, ",")
			wanted := make(map[string]bool)
			for _, p := range patterns {
				p = strings.TrimSpace(p)
				p = strings.TrimSuffix(p, ".")
				prefix := p + "."
				for _, n := range allNames {
					if n == p || strings.HasPrefix(n, prefix) {
						wanted[n] = true
					}
				}
			}

			// 过滤 doc.methods
			var filtered []json.RawMessage
			for _, raw := range all.Methods {
				if wanted[raw.Name] {
					// 保留原始格式
					b, _ := json.Marshal(raw)
					filtered = append(filtered, json.RawMessage(b))
				}
			}
			result := struct {
				OpenRPC string              `json:"openrpc"`
				Methods []json.RawMessage   `json:"methods"`
			}{OpenRPC: "1.4.1", Methods: filtered}
			b, _ := json.Marshal(result)
			writeRPC(w, rpcResponse{Result: json.RawMessage(b), ID: req.ID})
			return
		}

		result, err := reg.Execute(r.Context(), req.Method, req.Params)
		if err != nil {
			rpcErr := &agent.RPCError{Code: -32000, Message: err.Error()}
			if e, ok := err.(*agent.RPCError); ok {
				rpcErr = e
			}
			slog.Warn("rpc: method error", "method", req.Method, "err", rpcErr.Message)
			writeRPC(w, rpcResponse{Error: rpcErr, ID: req.ID})
			return
		}

		writeRPC(w, rpcResponse{Result: result, ID: req.ID})
	}
}

func writeRPC(w http.ResponseWriter, resp rpcResponse) {
	resp.JSONRPC = "2.0"
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Warn("rpc: write error", "err", err)
	}
}

