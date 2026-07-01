package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ffa/lan-party/internal/headscale"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) *Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	hs := headscale.New(srv.URL, "test-key")
	return New(hs)
}

func TestCreateParty(t *testing.T) {
	userCreated := false
	s := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/user" && r.Method == "POST" {
			userCreated = true
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]any{
					"id":   1,
					"name": "party-test",
				},
			})
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
	})

	body := `{"game":"Age of Empires II"}`
	req := httptest.NewRequest("POST", "/parties", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if !userCreated {
		t.Error("expected Headscale user to be created")
	}

	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["game"] != "Age of Empires II" {
		t.Errorf("unexpected game: %v", resp["game"])
	}
	if resp["id"] == nil || resp["id"] == "" {
		t.Error("expected non-empty party ID")
	}
}

func TestCreateInvite(t *testing.T) {
	s := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/user" && r.Method == "POST":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]any{"id": 1, "name": "party-test"},
			})
		case r.URL.Path == "/api/v1/preauthkey" && r.Method == "POST":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"preAuthKey": map[string]any{
					"id":       7,
					"key":      "invite-key-123",
					"reusable":  true,
				},
			})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	createBody := `{"game":"Quake III"}`
	req := httptest.NewRequest("POST", "/parties", bytes.NewBufferString(createBody))
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	var partyResp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &partyResp)
	partyID := partyResp["id"].(string)

	req2 := httptest.NewRequest("POST", "/parties/"+partyID+"/invite", nil)
	rec2 := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var invite map[string]any
	_ = json.Unmarshal(rec2.Body.Bytes(), &invite)
	if invite["code"] != "invite-key-123" {
		t.Errorf("unexpected code: %v", invite["code"])
	}
}

func TestGetParty(t *testing.T) {
	var partyID string
	s := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/user" && r.Method == "POST":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			partyID = body["name"].(string)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]any{"id": 1, "name": partyID},
			})
		case r.URL.Path == "/api/v1/node" && r.Method == "GET":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"nodes": []map[string]any{
					{
						"id":          1,
						"givenName":   "alice.tailnet",
						"ipAddresses": []string{"100.64.0.1"},
						"online":      true,
						"user":        map[string]any{"id": 1, "name": partyID},
					},
					{
						"id":          2,
						"givenName":   "bob.tailnet",
						"ipAddresses": []string{"100.64.0.2"},
						"online":      false,
						"user":        map[string]any{"id": 99, "name": "someone-else"},
					},
				},
			})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	req := httptest.NewRequest("POST", "/parties", bytes.NewBufferString(`{"game":"UT2004"}`))
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	var partyResp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &partyResp)

	req2 := httptest.NewRequest("GET", "/parties/"+partyID, nil)
	rec2 := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var resp map[string]any
	_ = json.Unmarshal(rec2.Body.Bytes(), &resp)
	players := resp["players"].([]any)
	if len(players) != 1 {
		t.Fatalf("expected 1 player (filtered), got %d", len(players))
	}
	player := players[0].(map[string]any)
	if player["ip"] != "100.64.0.1" {
		t.Errorf("unexpected IP: %v", player["ip"])
	}
}

func TestGetPartyNotFound(t *testing.T) {
	s := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("should not call headscale for unknown party")
	})

	req := httptest.NewRequest("GET", "/parties/nonexistent", nil)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestSetHost(t *testing.T) {
	s := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/user" && r.Method == "POST" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]any{"id": 1, "name": "party-test"},
			})
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
	})

	req := httptest.NewRequest("POST", "/parties", bytes.NewBufferString(`{"game":"Halo"}`))
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	var partyResp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &partyResp)
	partyID := partyResp["id"].(string)

	req2 := httptest.NewRequest("POST", "/parties/"+partyID+"/host", bytes.NewBufferString(`{"node_id":"node-42"}`))
	rec2 := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var resp map[string]any
	_ = json.Unmarshal(rec2.Body.Bytes(), &resp)
	if resp["host_node"] != "node-42" {
		t.Errorf("expected host_node node-42, got %v", resp["host_node"])
	}
}

func TestDeleteParty(t *testing.T) {
	userDeleted := false
	nodeDeleted := false
	keyExpired := false
	var partyID string

	s := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/user" && r.Method == "POST":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			partyID = body["name"].(string)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]any{"id": 1, "name": partyID},
			})
		case r.URL.Path == "/api/v1/preauthkey" && r.Method == "POST":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"preAuthKey": map[string]any{"id": 7, "key": "key123"},
			})
		case r.URL.Path == "/api/v1/node" && r.Method == "GET":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"nodes": []map[string]any{
					{
						"id":   1,
						"user": map[string]any{"id": 1, "name": partyID},
					},
				},
			})
		case r.URL.Path == "/api/v1/preauthkey/expire" && r.Method == "POST":
			keyExpired = true
			_ = json.NewEncoder(w).Encode(map[string]any{})
		case r.URL.Path == "/api/v1/node/1" && r.Method == "DELETE":
			nodeDeleted = true
			_ = json.NewEncoder(w).Encode(map[string]any{})
		case r.URL.Path == "/api/v1/user/1" && r.Method == "DELETE":
			userDeleted = true
			_ = json.NewEncoder(w).Encode(map[string]any{})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	req := httptest.NewRequest("POST", "/parties", bytes.NewBufferString(`{"game":"test"}`))
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)
	var partyResp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &partyResp)
	pid := partyResp["id"].(string)

	req2 := httptest.NewRequest("POST", "/parties/"+pid+"/invite", nil)
	rec2 := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec2, req2)

	req3 := httptest.NewRequest("DELETE", "/parties/"+pid, nil)
	rec3 := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec3.Code, rec3.Body.String())
	}
	if !keyExpired {
		t.Error("expected preauth key to be expired")
	}
	if !nodeDeleted {
		t.Error("expected node to be deleted")
	}
	if !userDeleted {
		t.Error("expected user to be deleted")
	}
}

func TestHealth(t *testing.T) {
	s := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("should not call headscale for health check")
	})

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %v", resp["status"])
	}
}
