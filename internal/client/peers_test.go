package client

import (
	"testing"
)

func TestParsePeers(t *testing.T) {
	data := []byte(`{
		"Self": {
			"HostName": "myhost",
			"TailscaleIPs": ["100.64.0.1", "fd7a:115c:a1e0::1"],
			"Online": true,
			"OS": "linux"
		},
		"Peer": {
			"nodeKey1": {
				"HostName": "alice",
				"TailscaleIPs": ["100.64.0.2"],
				"Online": true,
				"OS": "windows"
			},
			"nodeKey2": {
				"HostName": "bob",
				"TailscaleIPs": ["100.64.0.3"],
				"Online": false,
				"OS": "linux"
			}
		}
	}`)

	peers, err := parsePeers(data)
	if err != nil {
		t.Fatalf("parsePeers failed: %v", err)
	}

	if len(peers) != 3 {
		t.Fatalf("expected 3 peers, got %d", len(peers))
	}

	if peers[0].HostName != "myhost" || !peers[0].Online {
		t.Errorf("unexpected self: %+v", peers[0])
	}
	if peers[1].HostName != "alice" || !peers[1].Online {
		t.Errorf("unexpected peer 1: %+v", peers[1])
	}
	if peers[2].HostName != "bob" || peers[2].Online {
		t.Errorf("unexpected peer 2: %+v", peers[2])
	}
}

func TestFirstIPv4(t *testing.T) {
	p := Peer{
		TailScaleIPs: []string{"fd7a:115c:a1e0::1", "100.64.0.5"},
	}
	if p.FirstIPv4() != "100.64.0.5" {
		t.Errorf("expected 100.64.0.5, got %s", p.FirstIPv4())
	}

	p2 := Peer{
		TailScaleIPs: []string{"100.64.0.10"},
	}
	if p2.FirstIPv4() != "100.64.0.10" {
		t.Errorf("expected 100.64.0.10, got %s", p2.FirstIPv4())
	}

	p3 := Peer{
		TailScaleIPs: []string{},
	}
	if p3.FirstIPv4() != "" {
		t.Errorf("expected empty string, got %s", p3.FirstIPv4())
	}
}

func TestParsePeersInvalidJSON(t *testing.T) {
	_, err := parsePeers([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
