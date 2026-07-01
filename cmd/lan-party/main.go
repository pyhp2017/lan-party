package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/ffa/lan-party/internal/client"
	"github.com/ffa/lan-party/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "host":
		hostCmd(os.Args[2:])
	case "join":
		joinCmd(os.Args[2:])
	case "status":
		statusCmd(os.Args[2:])
	case "leave":
		leaveCmd(os.Args[2:])
	case "version":
		fmt.Println(version.Version)
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `lan-party - virtual LAN party tool

Usage:
  lan-party host [--game <name>] --server <url> --login-server <url>
  lan-party join <code> --server <url> --login-server <url>
  lan-party status [--server <url>]
  lan-party leave
  lan-party version

Flags:
  --server        lanpartyd control URL (e.g. http://localhost:8090)
  --login-server  Headscale URL (e.g. https://lanparty.example.com)
  --game          Game name (optional, for display)
  --party         Party ID (for status)

Environment:
  LANPARTY_SERVER_URL     Default for --server
  LANPARTY_LOGIN_SERVER   Default for --login-server`)
}

func commonFlags(fs *flag.FlagSet) (*string, *string) {
	server := fs.String("server", env("LANPARTY_SERVER_URL", ""), "lanpartyd control URL")
	loginServer := fs.String("login-server", env("LANPARTY_LOGIN_SERVER", ""), "Headscale login server URL")
	return server, loginServer
}

func hostCmd(args []string) {
	fs := flag.NewFlagSet("host", flag.ExitOnError)
	server, loginServer := commonFlags(fs)
	game := fs.String("game", "", "game name (optional)")
	_ = fs.Parse(args)

	if *server == "" {
		fmt.Fprintln(os.Stderr, "error: --server is required (or set LANPARTY_SERVER_URL)")
		os.Exit(1)
	}
	if *loginServer == "" {
		fmt.Fprintln(os.Stderr, "error: --login-server is required (or set LANPARTY_LOGIN_SERVER)")
		os.Exit(1)
	}

	gameName := strings.TrimSpace(*game)
	if gameName == "" {
		gameName = "untitled"
	}

	if !client.IsAdmin() {
		fmt.Fprintf(os.Stderr, "warning: admin/root recommended. %s\n", client.AdminInstructions())
	}

	c := client.New(*server, *loginServer)
	if err := ensureTailscale(c); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if runtime.GOOS == "windows" {
		_ = client.SetWindowsLoginServer(*loginServer)
	}

	body := fmt.Sprintf(`{"game":%q}`, gameName)
	resp, err := httpPost(*server+"/parties", body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create party:", err)
		os.Exit(1)
	}

	var party struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(resp, &party); err != nil {
		fmt.Fprintln(os.Stderr, "failed to parse response:", err)
		os.Exit(1)
	}

	fmt.Printf("Party created! ID: %s\n", party.ID)

	invite, err := httpPost(*server+"/parties/"+party.ID+"/invite", "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create invite:", err)
		os.Exit(1)
	}

	var inv struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(invite, &inv); err != nil {
		fmt.Fprintln(os.Stderr, "failed to parse invite:", err)
		os.Exit(1)
	}

	if err := c.TailscaleUp(inv.Code); err != nil {
		fmt.Fprintln(os.Stderr, "failed to join network:", err)
		os.Exit(1)
	}

	ip, err := c.SelfIP()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get own IP:", err)
		os.Exit(1)
	}

	_, _ = httpPost(*server+"/parties/"+party.ID+"/host", `{"node_id":"self"}`)

	fmt.Printf("\n")
	fmt.Printf("  =========================================\n")
	fmt.Printf("    Party ready! Share with friends:\n")
	fmt.Printf("  =========================================\n")
	fmt.Printf("\n")
	fmt.Printf("  Invite code: %s\n", inv.Code)
	fmt.Printf("  Your IP:     %s\n", ip)
	fmt.Printf("  Game:        %s\n", gameName)
	fmt.Printf("\n")
	fmt.Printf("  Players join with:\n")
	fmt.Printf("    lan-party join %s --server %s --login-server %s\n", inv.Code, *server, *loginServer)
	fmt.Printf("\n")
}

func joinCmd(args []string) {
	fs := flag.NewFlagSet("join", flag.ExitOnError)
	server, loginServer := commonFlags(fs)
	_ = fs.Parse(args)

	if *server == "" {
		fmt.Fprintln(os.Stderr, "error: --server is required (or set LANPARTY_SERVER_URL)")
		os.Exit(1)
	}
	if *loginServer == "" {
		fmt.Fprintln(os.Stderr, "error: --login-server is required (or set LANPARTY_LOGIN_SERVER)")
		os.Exit(1)
	}

	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "error: invite code required")
		fmt.Fprintln(os.Stderr, "usage: lan-party join <code> --server <url> --login-server <url>")
		os.Exit(1)
	}

	code := rest[0]

	if !client.IsAdmin() {
		fmt.Fprintf(os.Stderr, "warning: admin/root recommended. %s\n", client.AdminInstructions())
	}

	c := client.New(*server, *loginServer)
	if err := ensureTailscale(c); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if runtime.GOOS == "windows" {
		_ = client.SetWindowsLoginServer(*loginServer)
	}

	if err := c.TailscaleUp(code); err != nil {
		fmt.Fprintln(os.Stderr, "failed to join network:", err)
		os.Exit(1)
	}

	ip, err := c.SelfIP()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get own IP:", err)
		os.Exit(1)
	}

	fmt.Printf("Joined! Your IP: %s\n", ip)
	fmt.Printf("Ask the host for their IP, then connect in-game.\n")
}

func statusCmd(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	server, _ := commonFlags(fs)
	partyID := fs.String("party", "", "party ID (optional, shows party members)")
	_ = fs.Parse(args)

	c := client.New("", "")
	peers, err := c.Peers()
	if err != nil {
		fmt.Fprintln(os.Stderr, "not connected to a party:", err)
		os.Exit(1)
	}

	if *partyID != "" && *server != "" {
		resp, err := httpGet(*server + "/parties/" + *partyID)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to get party info:", err)
		} else {
			var pr struct {
				Party struct {
					Game string `json:"game"`
					HostIP string `json:"host_ip"`
				} `json:"party"`
				Players []struct {
					Name   string `json:"name"`
					IP     string `json:"ip"`
					Online bool   `json:"online"`
				} `json:"players"`
			}
			if err := json.Unmarshal(resp, &pr); err == nil {
				fmt.Printf("Party: %s\n", *partyID)
				if pr.Party.Game != "" {
					fmt.Printf("Game: %s\n", pr.Party.Game)
				}
				if pr.Party.HostIP != "" {
					fmt.Printf("Host IP: %s\n", pr.Party.HostIP)
				}
				fmt.Printf("\nPlayers (%d):\n", len(pr.Players))
				for _, p := range pr.Players {
					status := "offline"
					if p.Online {
						status = "online"
					}
					fmt.Printf("  %-20s %-16s %s\n", p.Name, p.IP, status)
				}
				fmt.Println()
			}
		}
	}

	fmt.Println("Tailscale peers:")
	for _, p := range peers {
		ip := p.FirstIPv4()
		status := "offline"
		if p.Online {
			status = "online"
		}
		fmt.Printf("  %-20s %-16s %s [%s]\n", p.HostName, ip, status, p.OS)
	}
}

func leaveCmd(_ []string) {
	c := client.New("", "")
	if err := c.TailscaleDown(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to leave:", err)
		os.Exit(1)
	}
	fmt.Println("Left the party.")
}

func ensureTailscale(c *client.Client) error {
	if _, err := c.DetectTailscale(); err != nil {
		return fmt.Errorf("tailscale not found\n\nInstall Tailscale first:\n  %s", client.TailscaleInstallInstructions())
	}
	return nil
}

func httpPost(url, body string) ([]byte, error) {
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(resp.Body)
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(resp.Body)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
