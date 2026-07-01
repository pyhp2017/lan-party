package headscale

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/user" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "party-test" {
			t.Errorf("unexpected name: %v", body["name"])
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"user": map[string]any{
				"id":   42,
				"name": "party-test",
			},
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "test-key")
	user, err := c.CreateUser("party-test")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if user.ID != 42 {
		t.Errorf("expected ID 42, got %d", user.ID)
	}
	if user.Name != "party-test" {
		t.Errorf("expected name party-test, got %s", user.Name)
	}
}

func TestCreatePreAuthKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/preauthkey" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["user"] != float64(42) {
			t.Errorf("expected user 42, got %v", body["user"])
		}
		if body["reusable"] != true {
			t.Errorf("expected reusable true, got %v", body["reusable"])
		}
		if body["ephemeral"] != false {
			t.Errorf("expected ephemeral false, got %v", body["ephemeral"])
		}
		if body["expiration"] == nil {
			t.Error("expected expiration to be set")
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"preAuthKey": map[string]any{
				"id":       7,
				"key":      "abc123def456",
				"reusable":  true,
				"ephemeral": false,
			},
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "test-key")
	key, err := c.CreatePreAuthKey(42, true, false, time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("CreatePreAuthKey failed: %v", err)
	}
	if key.ID != 7 {
		t.Errorf("expected ID 7, got %d", key.ID)
	}
	if key.Key != "abc123def456" {
		t.Errorf("expected key abc123def456, got %s", key.Key)
	}
}

func TestListNodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/node" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("unexpected method: %s", r.Method)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"nodes": []map[string]any{
				{
					"id":          1,
					"machineKey":  "mk1",
					"nodeKey":     "nk1",
					"ipAddresses": []string{"100.64.0.1", "fd7a:115c:a1e0::1"},
					"name":        "host1",
					"givenName":   "host1.tailnet",
					"online":      true,
					"user": map[string]any{
						"id":   42,
						"name": "party-test",
					},
				},
				{
					"id":         2,
					"givenName":  "host2.tailnet",
					"ipAddresses": []string{"100.64.0.2"},
					"online":     false,
					"user": map[string]any{
						"id":   99,
						"name": "other",
					},
				},
			},
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "test-key")
	nodes, err := c.ListNodes()
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].ID != 1 || nodes[0].IPAddresses[0] != "100.64.0.1" || !nodes[0].Online {
		t.Errorf("unexpected node[0]: %+v", nodes[0])
	}
	if nodes[1].ID != 2 || nodes[1].Online {
		t.Errorf("unexpected node[1]: %+v", nodes[1])
	}
}

func TestDeleteNode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/node/5" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer srv.Close()

	c := New(srv.URL, "test-key")
	if err := c.DeleteNode(5); err != nil {
		t.Fatalf("DeleteNode failed: %v", err)
	}
}

func TestExpirePreAuthKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/preauthkey/expire" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["id"] != float64(7) {
			t.Errorf("expected id 7, got %v", body["id"])
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer srv.Close()

	c := New(srv.URL, "test-key")
	if err := c.ExpirePreAuthKey(7); err != nil {
		t.Fatalf("ExpirePreAuthKey failed: %v", err)
	}
}

func TestAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"unauthorized"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "bad-key")
	_, err := c.CreateUser("test")
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}
