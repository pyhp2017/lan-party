# lan-party

Bring back early-2000s LAN parties over the internet. `lan-party` creates a
virtual LAN so old-school multiplayer games think everyone is on the same
local network — just type the host's IP and connect.

## How it works

- **Headscale** (self-hosted Tailscale control server, v0.29.1) runs on a VPS
  with an embedded DERP relay for NAT traversal.
- **lanpartyd** (Go) is the control plane: creates parties, mints invite codes
  (Headscale preauth keys), tracks who's the host.
- **lan-party** CLI is the player tool: `host` creates a party, `join <code>`
  brings the player onto the shared tailnet. Everyone gets a `100.64.x.y` IP.
- The official **Tailscale** client on each player's machine handles the actual
  WireGuard tunnel with automatic NAT traversal (STUN + DERP fallback).

```
                 ┌─────────────────────────┐
   VPS ────────> │ Headscale (v0.29.1)     │
                 │ + embedded DERP         │
                 │ + lanpartyd             │
                 │ + Caddy (TLS)           │
                 └───────────┬─────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        ▼                    ▼                    ▼
   Player A             Player B (host)      Player C
   lan-party join       lan-party host       lan-party join
   100.64.0.2           100.64.0.1           100.64.0.3
        \__________ WireGuard mesh ________/
```

## Quick start (server)

### 1. Provision the VPS

```bash
cd deploy/terraform
# Create terraform.tfvars:
#   hcloud_token  = "your-hetzner-api-token"
#   ssh_key_name  = "your-ssh-key-name-in-hetzner"
#   domain        = "lanparty.yourdomain.com"
terraform init
terraform apply
```

Cloud-init automatically installs Docker, Headscale, and Caddy.

### 2. Point DNS

Create an A record for your domain pointing to the VPS IP (shown in Terraform
output). Wait for DNS to propagate — Caddy will auto-issue a Let's Encrypt
certificate.

### 3. Generate Headscale API key

```bash
ssh root@<server-ip>
docker exec headscale headscale apikeys create
```

### 4. Start lanpartyd

```bash
git clone https://github.com/ffa/lan-party.git /opt/lan-party
cd /opt/lan-party
export HEADSCALE_API_KEY="<the-key-from-step-3>"
export HEADSCALE_URL="http://headscale:8080"
docker compose -f deploy/docker/docker-compose.yml up -d lanpartyd
```

## Quick start (player)

### Prerequisites

- Install the [Tailscale client](https://tailscale.com/download)
  - Windows: download the MSI installer (run as admin)
  - Linux: `curl -fsSL https://tailscale.com/install.sh | sh`
- Get an invite code from the party host
- Run as admin/root (required for Tailscale)

### Host a party

```bash
export LANPARTY_SERVER_URL=https://lanparty.yourdomain.com
export LANPARTY_LOGIN_SERVER=https://lanparty.yourdomain.com

lan-party host --game "Age of Empires II"
```

This prints an invite code and your IP on the virtual LAN.

### Join a party

```bash
export LANPARTY_SERVER_URL=https://lanparty.yourdomain.com
export LANPARTY_LOGIN_SERVER=https://lanparty.yourdomain.com

lan-party join <invite-code>
```

### Check who's online

```bash
lan-party status --party <party-id>
```

### Connect in-game

The host tells everyone their `100.64.x.y` IP. In the game, choose Direct IP
Connect / LAN and type that address.

### Leave

```bash
lan-party leave
```

## Build

```bash
make build          # both binaries
make cross-compile  # windows/amd64 + linux/amd64
make test           # unit tests
make lint           # golangci-lint
```

## Releases

Binaries are built and published automatically when a tag is pushed:

```bash
git tag v0.1.0
git push origin v0.1.0
```

This triggers the Release workflow, which cross-compiles and publishes:
- `lan-party-vX.Y.Z-linux-amd64.tar.gz` — player CLI (Linux)
- `lan-party-vX.Y.Z-windows-amd64.zip` — player CLI (Windows)
- `lanpartyd-vX.Y.Z-linux-amd64.tar.gz` — server (Linux VPS)
- `checksums.txt` — SHA256 sums

Download the latest release from the [Releases page](../../releases).

## Ports

| Port | Protocol | Purpose |
|------|----------|---------|
| 443  | TCP | HTTPS — Headscale API + DERP relay |
| 80   | TCP | ACME HTTP-01 challenge (Caddy auto-TLS) |
| 3478 | UDP | STUN — NAT traversal |

## Tech stack

- **Go** — orchestrator + CLI
- **Headscale v0.29.1** — self-hosted Tailscale control plane (SQLite, embedded DERP)
- **Caddy** — reverse proxy + automatic Let's Encrypt TLS
- **Terraform** — Hetzner Cloud VPS provisioning
- **Docker Compose** — server deployment
- **GitHub Actions** — CI + release automation

## Project structure

```
cmd/lan-party/        Player CLI
cmd/lanpartyd/        Server control plane
internal/headscale/   Headscale REST API client
internal/server/      HTTP handlers (parties, invites, host)
internal/client/      Tailscale wrapper (detect, up/down, peers)
internal/party/       Domain types
deploy/docker/        docker-compose + Headscale config + Caddyfile
deploy/terraform/     Hetzner VPS + cloud-init
.github/workflows/    CI + release workflows
```
