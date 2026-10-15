// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package table

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTableDataBatchCreate_ProxyArrayAndResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gw/ai/proxy" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var req struct {
			Path   string                   `json:"path"`
			Method string                   `json:"method"`
			Body   []map[string]interface{} `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		if req.Path != "/basicdata/record/batchSave" || req.Method != "POST" {
			t.Errorf("unexpected proxy target: %s %s", req.Method, req.Path)
		}
		if len(req.Body) != 2 {
			t.Errorf("expected 2 records, got %d", len(req.Body))
			return
		}
		for i, rec := range req.Body {
			if rec["basicDataId"] != float64(1633) || rec["schemaId"] != float64(2546) || rec["version"] != float64(5) {
				t.Errorf("record %d missing shared values: %v", i, rec)
			}
			if _, ok := rec["tenantId"]; ok {
				t.Errorf("record %d should not contain tenantId", i)
			}
		}
		if req.Body[0]["data"].(map[string]interface{})["1"] != "产品A" || req.Body[1]["data"].(map[string]interface{})["2"] != float64(20) {
			t.Errorf("data was not preserved: %v", req.Body)
		}
		writeProxyResponse(t, w, map[string]interface{}{"success": true, "result": []interface{}{map[string]interface{}{"id": 101}, map[string]interface{}{"id": 102}}})
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-create", "--basic-data-id", "1633", "--schema-id", "2546", "--version", "5", "--records", `[{"data":{"1":"产品A"}},{"data":{"1":"产品B","2":20}}]`})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out.String(), `"id": 101`) {
		t.Fatalf("missing returned records: %s", out.String())
	}
}

func TestTableDataBatchCreate_RecordsFileAndPerRecordVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "records.json")
	if err := os.WriteFile(path, []byte(`[{"data":{"1":"A"}},{"data":{"1":"B"},"version":9}]`), 0600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Body []map[string]interface{} `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		if len(req.Body) != 2 || req.Body[1]["version"] != float64(9) {
			t.Errorf("file input or per-record version lost: %v", req.Body)
		}
		if _, ok := req.Body[0]["version"]; ok {
			t.Errorf("unspecified version must be omitted: %v", req.Body[0])
		}
		writeProxyResponse(t, w, map[string]interface{}{"success": true, "result": req.Body})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-create", "--basic-data-id", "1633", "--schema-id", "2546", "--records-file", path})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
}

func TestTableDataBatchCreate_BackendFailureReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeProxyResponse(t, w, map[string]interface{}{
			"success": false,
			"msg":     "database busy",
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-create", "--basic-data-id", "1633", "--schema-id", "2546", "--records", `[{"data":{"1":"A"}}]`})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "database busy") {
		t.Fatalf("expected backend failure, got %v", err)
	}
}

func TestTableDataBatchCreate_ResultCountMismatchReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  []interface{}{map[string]interface{}{"id": 101}},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-create", "--basic-data-id", "1633", "--schema-id", "2546", "--records", `[{"data":{"1":"A"}},{"data":{"1":"B"}}]`})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "returned 1 records, expected 2") {
		t.Fatalf("expected result count mismatch, got %v", err)
	}
}

func TestTableDataBatchCreate_InvalidInputDoesNotSend(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		writeProxyResponse(t, w, map[string]interface{}{"success": true})
	}))
	defer server.Close()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"missing source", nil, "records"},
		{"both sources", []string{"--records", `[{"data":{"1":"A"}}]`, "--records-file", "missing.json"}, "only one"},
		{"not array", []string{"--records", `{"data":{"1":"A"}}`}, "invalid --records"},
		{"empty", []string{"--records", `[]`}, "at least one"},
		{"null entry", []string{"--records", `[null]`}, "record 1"},
		{"empty data", []string{"--records", `[{"data":{}}]`}, "record 1"},
		{"mixed table", []string{"--records", `[{"basicDataId":999,"data":{"1":"A"}}]`}, "basicDataId"},
		{"mixed schema", []string{"--records", `[{"schemaId":999,"data":{"1":"A"}}]`}, "schemaId"},
		{"missing file", []string{"--records-file", "missing.json"}, "read --records-file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _, errOut := newTestFactory(server.URL)
			cmd := NewCmdTable(f)
			cmd.SetErr(errOut)
			args := append([]string{"data", "batch-create", "--basic-data-id", "1633", "--schema-id", "2546"}, tt.args...)
			cmd.SetArgs(args)
			if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q error, got %v", tt.want, err)
			}
		})
	}
	if requests != 0 {
		t.Fatalf("invalid input sent %d requests", requests)
	}
}
