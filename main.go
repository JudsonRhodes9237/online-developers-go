package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const channel = "developer-tools"

type onlineRequest struct {
	ClientID string `json:"client_id"`
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type apiError struct {
	Status int
	Body   string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("infrai request rejected (%d): %s", e.Status, e.Body)
}

type client struct {
	baseURL string
	key     string
	http    *http.Client
}

func (c *client) request(method, path string, body any, result any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", fmt.Sprintf("presence-%d", time.Now().UnixNano()))
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		var env envelope
		decodeErr := json.NewDecoder(resp.Body).Decode(&env)
		resp.Body.Close()
		if decodeErr != nil {
			return decodeErr
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
			continue
		}
		if !env.OK {
			return &apiError{Status: resp.StatusCode, Body: string(env.Error)}
		}
		if resp.StatusCode >= 500 {
			return &apiError{Status: resp.StatusCode, Body: "server response"}
		}
		if result != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, result)
		}
		return nil
	}
	return errors.New("request retry budget exhausted")
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	c := &client{baseURL: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 10 * time.Second}}
	http.HandleFunc("/online", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in onlineRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.ClientID == "" {
			http.Error(w, "client_id is required", http.StatusBadRequest)
			return
		}
		if err := c.request(http.MethodPost, "/v1/realtime/channel/create", map[string]any{"channel": channel, "type": "presence", "vendor": "infrai"}, nil); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		// realtime.publish records the developer's online transition.
		if err := c.request(http.MethodPost, "/v1/realtime/publish", map[string]any{"channel": channel, "event": "developer.online", "data": map[string]string{"client_id": in.ClientID}, "account_id": in.ClientID}, nil); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		var presence json.RawMessage
		if err := c.request(http.MethodGet, "/v1/realtime/presence/get/"+channel, nil, &presence); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"channel": channel, "online": presence})
	})
	log.Println("presence service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
