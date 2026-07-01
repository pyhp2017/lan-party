package party

import "time"

type Party struct {
	ID           string    `json:"id"`
	HeadscaleUID  uint64   `json:"-"`
	Game         string    `json:"game"`
	HostNode     string    `json:"host_node,omitempty"`
	HostIP       string    `json:"host_ip,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Invite struct {
	Code       string    `json:"code"`
	PreAuthKey string    `json:"-"`
	KeyID      uint64    `json:"-"`
	ExpiresAt  time.Time `json:"expires_at"`
	Used       bool      `json:"used"`
}

type Player struct {
	NodeID   string    `json:"node_id"`
	Name     string    `json:"name"`
	IP       string    `json:"ip"`
	Online   bool      `json:"online"`
	JoinedAt time.Time `json:"joined_at"`
}

type CreatePartyRequest struct {
	Game string `json:"game"`
}

type JoinRequest struct {
	Code string `json:"code"`
}

type SetHostRequest struct {
	NodeID string `json:"node_id"`
}

type PartyResponse struct {
	Party   Party    `json:"party"`
	Players []Player `json:"players"`
}
