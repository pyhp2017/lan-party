output "server_ip" {
  description = "Public IPv4 of the VPS"
  value       = hcloud_server.lanparty.ipv4_address
}

output "server_ipv6" {
  description = "Public IPv6 of the VPS"
  value       = hcloud_server.lanparty.ipv6_address
}

output "domain" {
  description = "Domain configured for Headscale (create DNS A record pointing to server_ip)"
  value       = var.domain
}

output "ssh_command" {
  description = "SSH into the VPS"
  value       = "ssh root@${hcloud_server.lanparty.ipv4_address}"
}
