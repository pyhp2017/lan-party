variable "hcloud_token" {
  description = "Hetzner Cloud API token"
  type        = string
  sensitive   = true
}

variable "ssh_key_name" {
  description = "Name of an existing Hetzner SSH public key to use for the server"
  type        = string
}

variable "server_name" {
  description = "Name of the VPS"
  type        = string
  default     = "lanparty"
}

variable "server_type" {
  description = "Hetzner server type (CX22 = 2 vCPU, 4GB RAM)"
  type        = string
  default     = "cx22"
}

variable "image" {
  description = "OS image for the VPS"
  type        = string
  default     = "ubuntu-24.04"
}

variable "location" {
  description = "Hetzner datacenter location (nbg1, fsn1, hel1, ash, hil, sin)"
  type        = string
  default     = "nbg1"
}

variable "domain" {
  description = "Domain name for Headscale (must have DNS A record pointing to the server IP)"
  type        = string
  default     = "lanparty.example.com"
}
