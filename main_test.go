package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequestBodyEndsAtEOF(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request: %v", err)
			return
		}
		var got map[string]string
		if err := json.Unmarshal(body, &got); err != nil || got["channel"] != "developer-tools" {
			t.Errorf("request body = %s, error = %v", body, err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"data":{}}`))
	}))
	defer server.Close()
	c := &client{baseURL: server.URL, key: "test", http: &http.Client{Timeout: time.Second}}
	if err := c.request(http.MethodPost, "/v1/realtime/channel/create", map[string]string{"channel": "developer-tools"}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestOnlineRequestNeedsClientID(t *testing.T) {
	cases := []struct {
		name, id string
		valid    bool
	}{{"cli identity", "alice", true}, {"empty", "", false}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if (tc.id != "") != tc.valid {
				t.Fatalf("client id validity = %v", tc.valid)
			}
		})
	}
}
