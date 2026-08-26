package http

import (
	"bytes"
	"encoding/json"
	"errors"
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
		// CORS 头由 CORSMiddleware 统一管理（允许来源白名单：liki.hk / dev localhost）——
		// 此处不得覆盖为 "*"（曾无条件设 * 使白名单形同虚设，任意来源可跨域调用）
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
			var mbe *http.MaxBytesError
			if errors.As(err, &mbe) {
				writeRPC(w, rpcResponse{Error: &agent.RPCError{Code: -32600, Message: "request body too large (max 1MB)"}, ID: nil})
				return
			}
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
			var fp struct {
				Methods string `json:"methods,omitempty"`
			}
			_ = json.Unmarshal(req.Params, &fp) // 解析失败按无 methods 过滤处理（rpc.discover 降级）

			var patterns []string
			if fp.Methods != "" {
				patterns = strings.Split(fp.Methods, ",")
			}
			writeRPC(w, rpcResponse{Result: reg.DiscoverMethods(patterns), ID: req.ID})
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
