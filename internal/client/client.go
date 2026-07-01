package client

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Client struct {
	serverURL   string
	loginServer string
}

func New(serverURL, loginServer string) *Client {
	return &Client{
		serverURL:   serverURL,
		loginServer: loginServer,
	}
}

func (c *Client) DetectTailscale() (string, error) {
	cmd := exec.Command("tailscale", "version")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tailscale not found: %w", err)
	}
	line := strings.SplitN(string(out), "\n", 2)[0]
	return line, nil
}

func (c *Client) TailscaleUp(authKey string) error {
	args := []string{
		"up",
		"--login-server", c.loginServer,
		"--authkey", authKey,
		"--accept-dns=true",
	}
	cmd := exec.Command("tailscale", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tailscale up failed: %w: %s", err, string(out))
	}
	return nil
}

func (c *Client) TailscaleDown() error {
	cmd := exec.Command("tailscale", "down")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tailscale down failed: %w: %s", err, string(out))
	}
	return nil
}

func (c *Client) Status() (string, error) {
	cmd := exec.Command("tailscale", "status")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tailscale status failed: %w", err)
	}
	return string(out), nil
}

func (c *Client) SelfIP() (string, error) {
	cmd := exec.Command("tailscale", "ip", "-4")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tailscale ip failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *Client) Peers() ([]Peer, error) {
	cmd := exec.Command("tailscale", "status", "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("tailscale status --json failed: %w", err)
	}
	return parsePeers(out)
}

func SetWindowsLoginServer(loginServer string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("only supported on Windows")
	}
	key := `HKLM\SOFTWARE\Tailscale IPN\Settings`
	cmd := exec.Command("reg", "add", key, "/v", "LoginURL", "/t", "REG_SZ", "/d", loginServer, "/f")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set registry key: %w: %s", err, string(out))
	}
	return nil
}

func TailscaleInstallInstructions() string {
	switch runtime.GOOS {
	case "windows":
		return "Download and install from https://tailscale.com/download/windows"
	case "linux":
		return "Install via your distro's package manager. See https://tailscale.com/download/linux"
	case "darwin":
		return "brew install tailscale"
	default:
		return "See https://tailscale.com/download"
	}
}

func IsAdmin() bool {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("net", "session")
		return cmd.Run() == nil
	case "linux", "darwin":
		return os.Geteuid() == 0
	default:
		return false
	}
}

func AdminInstructions() string {
	switch runtime.GOOS {
	case "windows":
		return "Right-click the terminal and select 'Run as administrator'."
	case "linux", "darwin":
		return "Re-run with sudo."
	default:
		return "Run as an administrator."
	}
}
