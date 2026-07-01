package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ffa/lan-party/internal/headscale"
	"github.com/ffa/lan-party/internal/party"
)

type Server struct {
	hs      *headscale.Client
	mu      sync.RWMutex
	parties map[string]*partyState
}

type partyState struct {
	party      party.Party
	inviteKeys []uint64
}

func New(hs *headscale.Client) *Server {
	return &Server{
		hs:      hs,
		parties: make(map[string]*partyState),
	}
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /parties", s.createParty)
	mux.HandleFunc("GET /parties/{id}", s.getParty)
	mux.HandleFunc("POST /parties/{id}/invite", s.createInvite)
	mux.HandleFunc("POST /parties/{id}/host", s.setHost)
	mux.HandleFunc("DELETE /parties/{id}", s.deleteParty)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) createParty(w http.ResponseWriter, r *http.Request) {
	var req party.CreatePartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	partyID := generateID()

	user, err := s.hs.CreateUser(partyID)
	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("failed to create headscale user: %v", err))
		return
	}

	p := party.Party{
		ID:           partyID,
		HeadscaleUID:  user.ID,
		Game:          req.Game,
		CreatedAt:     time.Now(),
	}

	s.mu.Lock()
	s.parties[partyID] = &partyState{party: p}
	s.mu.Unlock()

	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) getParty(w http.ResponseWriter, r *http.Request) {
	partyID := r.PathValue("id")

	s.mu.RLock()
	ps, ok := s.parties[partyID]
	s.mu.RUnlock()

	if !ok {
		writeError(w, http.StatusNotFound, "party not found")
		return
	}

	nodes, err := s.hs.ListNodes()
	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("failed to list nodes: %v", err))
		return
	}

	players := make([]party.Player, 0, len(nodes))
	for _, n := range nodes {
		if n.User.Name != partyID {
			continue
		}
		ip := ""
		if len(n.IPAddresses) > 0 {
			ip = n.IPAddresses[0]
		}
		players = append(players, party.Player{
			NodeID:   fmt.Sprintf("%d", n.ID),
			Name:     n.GivenName,
			IP:       ip,
			Online:   n.Online,
			JoinedAt: time.Now(),
		})
	}

	writeJSON(w, http.StatusOK, party.PartyResponse{
		Party:   ps.party,
		Players: players,
	})
}

func (s *Server) createInvite(w http.ResponseWriter, r *http.Request) {
	partyID := r.PathValue("id")

	s.mu.RLock()
	ps, ok := s.parties[partyID]
	s.mu.RUnlock()

	if !ok {
		writeError(w, http.StatusNotFound, "party not found")
		return
	}

	expiry := time.Now().Add(24 * time.Hour)
	key, err := s.hs.CreatePreAuthKey(ps.party.HeadscaleUID, true, false, expiry)
	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("failed to create invite: %v", err))
		return
	}

	s.mu.Lock()
	ps.inviteKeys = append(ps.inviteKeys, key.ID)
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, party.Invite{
		Code:       key.Key,
		PreAuthKey: key.Key,
		KeyID:      key.ID,
		ExpiresAt:  expiry,
	})
}

func (s *Server) setHost(w http.ResponseWriter, r *http.Request) {
	partyID := r.PathValue("id")

	var req party.SetHostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ps, ok := s.parties[partyID]
	if !ok {
		writeError(w, http.StatusNotFound, "party not found")
		return
	}

	ps.party.HostNode = req.NodeID

	writeJSON(w, http.StatusOK, ps.party)
}

func (s *Server) deleteParty(w http.ResponseWriter, r *http.Request) {
	partyID := r.PathValue("id")

	s.mu.Lock()
	defer s.mu.Unlock()

	ps, ok := s.parties[partyID]
	if !ok {
		writeError(w, http.StatusNotFound, "party not found")
		return
	}

	for _, keyID := range ps.inviteKeys {
		_ = s.hs.ExpirePreAuthKey(keyID)
	}

	nodes, err := s.hs.ListNodes()
	if err == nil {
		for _, n := range nodes {
			if n.User.Name == partyID {
				_ = s.hs.DeleteNode(n.ID)
			}
		}
	}

	_ = s.hs.DeleteUser(ps.party.HeadscaleUID)
	delete(s.parties, partyID)

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func generateID() string {
	return fmt.Sprintf("party-%d", time.Now().UnixNano())
}
