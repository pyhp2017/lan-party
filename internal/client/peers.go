package client

import (
	"encoding/json"
	"strings"
)

type statusJSON struct {
	Self  peerJSON            `json:"Self"`
	Peer  map[string]peerJSON `json:"Peer"`
}

type peerJSON struct {
	HostName     string   `json:"HostName"`
	TailScaleIPs []string `json:"TailscaleIPs"`
	Online       bool     `json:"Online"`
	OS           string   `json:"OS"`
}

type Peer struct {
	HostName     string
	TailScaleIPs []string
	Online       bool
	OS           string
}

func parsePeers(data []byte) ([]Peer, error) {
	var sj statusJSON
	if err := json.Unmarshal(data, &sj); err != nil {
		return nil, err
	}

	peers := make([]Peer, 0, len(sj.Peer)+1)
	peers = append(peers, Peer(sj.Self))

	for _, p := range sj.Peer {
		peers = append(peers, Peer(p))
	}

	return peers, nil
}

func (p Peer) FirstIPv4() string {
	for _, ip := range p.TailScaleIPs {
		if strings.Contains(ip, ".") {
			return ip
		}
	}
	if len(p.TailScaleIPs) > 0 {
		return p.TailScaleIPs[0]
	}
	return ""
}
