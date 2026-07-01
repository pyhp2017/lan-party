package client

import (
	"cmp"
	"slices"
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

	rest := peers[1:]
	slices.SortFunc(rest, func(a, b Peer) int {
		return cmp.Compare(a.HostName, b.HostName)
	})

	if rest[0].HostName != "alice" || !rest[0].Online || rest[0].OS != "windows" {
		t.Errorf("unexpected alice: %+v", rest[0])
	}
	if rest[1].HostName != "bob" || rest[1].Online || rest[1].OS != "linux" {
		t.Errorf("unexpected bob: %+v", rest[1])
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
