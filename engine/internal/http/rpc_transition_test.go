package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"liki-engine/internal/agent"
)

func TestHandleRPCDiscoverAndExecuteTransitionContract(t *testing.T) {
	t.Parallel()

	registry := agent.NewRPCRegistry()
	registry.SetVersion("2026.09.26.0")
	handler := HandleRPC(registry)

	discover := httptest.NewRequest(http.MethodPost, "/jsonrpc", strings.NewReader(`{"jsonrpc":"2.0","method":"rpc.discover","id":1}`))
	recorder := httptest.NewRecorder()
	handler(recorder, discover)
	if recorder.Code != http.StatusOK {
		t.Fatalf("discover status = %d, want 200", recorder.Code)
	}
	var discovery struct {
		Result struct {
			Info struct {
				Version string `json:"version"`
			} `json:"info"`
		} `json:"result"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &discovery); err != nil {
		t.Fatal(err)
	}
	if discovery.Result.Info.Version != "2026.09.26.0" {
		t.Fatalf("version = %q", discovery.Result.Info.Version)
	}

	body := `{"jsonrpc":"2.0","method":"time.now","params":{},"id":2}`
	request := httptest.NewRequest(http.MethodPost, "/jsonrpc", strings.NewReader(body))
	recorder = httptest.NewRecorder()
	handler(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("execute status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"cst"`) {
		t.Fatalf("execute body = %s, want time.now data", recorder.Body.String())
	}
}
