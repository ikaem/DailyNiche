# Raspberry Pi Home Server Setup

## Goal

Turn the Raspberry Pi 5 into a general-purpose personal home server - a substrate for deploying anything (Dockerized or not), not just DailyNiche. DailyNiche's staging deployment (CLAUDE.md's Phase 9) is the first thing to run on it, but the setup here is deliberately not DailyNiche-specific, so future projects can land on the same box without redoing this work.

## Starting Point (as of 2026-07-19)

- Raspberry Pi 5, hardware assembled, no OS installed yet
- microSD card for storage
- Headless setup planned - no monitor/keyboard, WiFi network
- Dev machine (this one) not yet confirmed to have an SSH keypair - check before Phase A

## Phase A: Flash & First Boot

- [ ] A.1: Check for an existing SSH keypair on the dev machine (`ls ~/.ssh/*.pub`); generate one if none exists (`ssh-keygen -t ed25519`). This key goes into the Pi's authorized_keys during imaging, so the Pi never needs password auth exposed even on its very first boot.
- [ ] A.2: Install Raspberry Pi Imager on the dev machine, if not already installed (apt/snap/flatpak, or direct download from raspberrypi.com).
- [ ] A.3: In Imager, choose OS = **Raspberry Pi OS Lite (64-bit)** - no desktop environment needed for a headless server, keeps resource usage (RAM/CPU/SD wear) low.
- [ ] A.4: Choose storage = the microSD card.
- [ ] A.5: Before writing, open the OS customization/advanced options (gear icon, or Ctrl+Shift+X) and configure:
  - Hostname - something general-purpose (e.g. `homelab` or `pi5`), not `dailyniche`, since this box is meant to outlive any single project
  - Enable SSH, "Allow public-key authentication only," paste the dev machine's public key from A.1
  - Username - avoid the old default `pi` (current Raspberry Pi OS no longer ships it by default, but confirm)
  - WiFi: SSID, password, country code (required for legal WiFi channel regulations)
  - Locale, timezone, keyboard layout
- [ ] A.6: Write the image to the microSD card.
- [ ] A.7: Insert the card into the Pi, connect power. Wait 1-2 minutes for first boot (WiFi association + SSH startup take a moment).

## Phase B: Confirm Access & Base Hardening

- [ ] B.1: From the dev machine: `ssh <username>@<hostname>.local` (relies on mDNS/Avahi, which Raspberry Pi OS ships with enabled by default). Fallback if `.local` doesn't resolve: check the router's connected-devices/DHCP client list for the Pi's IP.
- [ ] B.2: Once connected: `sudo apt update && sudo apt full-upgrade -y`; reboot if a kernel/firmware update was applied.
- [ ] B.3: Confirm password authentication is actually disabled, not just key auth enabled - check `/etc/ssh/sshd_config` for `PasswordAuthentication no` (Imager's "public-key only" option should set this; verify rather than assume).
- [ ] B.4: Decide on a static IP reservation for the Pi in the router's DHCP settings (or rely on hostname resolution) - avoids the Pi's address changing later and breaking Cloudflare Tunnel config or bookmarks.
- [ ] B.5 (optional, decide later): set up `unattended-upgrades` or at least a routine for periodic `apt upgrade`.

## Phase C: General-Purpose Deployment Substrate

- [ ] C.1: Install Docker + Docker Compose (Docker's official convenience install script works on Raspberry Pi OS's ARM64; confirm compose is the `docker compose` plugin form, not the deprecated standalone `docker-compose`).
- [ ] C.2: Add the Pi's user to the `docker` group, so Docker commands don't need `sudo` for every invocation.
- [ ] C.3: Install a reverse proxy - **Caddy** recommended (automatic HTTPS via Let's Encrypt, simplest config of the common options: Caddy/Traefik/nginx) - so multiple future services can each get their own hostname/path on one box.
- [ ] C.4: Install and configure Cloudflare Tunnel (`cloudflared`): authenticate, create a tunnel, route it through the Pi. This is the general mechanism for exposing *any* future service publicly without ever touching router port-forwarding - DailyNiche's Phase 9.2 (in CLAUDE.md) becomes just "add one more route to this tunnel," not a separate setup.
- [ ] C.5: Decide on a directory convention for future projects living on this box (e.g. `/srv/<project-name>/docker-compose.yml`), so unrelated services don't collide or get tangled together.

## Phase D: DailyNiche-Specific Deployment

Once Phases A-C are done, continue with CLAUDE.md's existing Phase 9 tasks, now layered on top of this general substrate instead of being a one-off setup:

- [ ] 9.1: Dockerfile + docker-compose.yml for DailyNiche specifically
- [ ] 9.3: Daily cron job on the Pi's host OS, invoking the fetcher via `docker-compose exec` (already decided 2026-07-18, see CLAUDE.md)
- [ ] 9.4 (optional): Backup strategy for the SQLite file

---

**Status:** Plan only - nothing in this document has been executed yet. Picking this up starts at Phase A, Task A.1.
