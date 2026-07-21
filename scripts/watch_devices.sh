#!/bin/bash
# Continuously scans a subnet and reports newly-appearing devices (anything
# that responds on port 22, open or closed - "closed" still proves a host is
# alive and reachable, just not running SSH).
#
# Useful on networks where a plain ping sweep is unreliable (e.g. client/AP
# isolation on phone hotspots blocking ICMP between clients even when direct
# TCP still gets through) - built while debugging exactly that scenario.
#
# A repeated, fast scan tolerates occasional missed passes (a miss just gets
# caught on the next pass a few seconds later), which is why this can safely
# use more aggressive nmap timing than a one-shot diagnostic scan would want.
#
# Usage: ./watch_devices.sh [subnet-in-CIDR]
#   If no subnet is given, auto-detects this machine's own address/prefix via
#   its default route (e.g. 10.198.140.109/24) and scans that.

set -euo pipefail

if [[ -n "${1:-}" ]]; then
  subnet="$1"
else
  iface=$(ip route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if ($i=="dev") print $(i+1)}')
  subnet=$(ip -4 addr show "$iface" | awk '/inet /{print $2; exit}')
fi

own_ip=$(ip route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if ($i=="src") print $(i+1)}')
gateway_ip=$(ip route | awk '/^default/ {print $3; exit}')

declare -A seen
[[ -n "$own_ip" ]] && seen["$own_ip"]=1
[[ -n "$gateway_ip" ]] && seen["$gateway_ip"]=1

echo "$(date '+%H:%M:%S') watching $subnet for new devices (pre-seeded: this machine=$own_ip, gateway=$gateway_ip)..."

while true; do
  hosts=$(nmap -sT -Pn -p22 -T4 --max-retries 1 -oG - "$subnet" 2>/dev/null \
    | grep -E "22/(open|closed)/tcp" \
    | awk '{print $2}')

  for ip in $hosts; do
    if [[ -z "${seen[$ip]:-}" ]]; then
      echo "$(date '+%H:%M:%S') NEW DEVICE: $ip"
      seen[$ip]=1
    fi
  done

  sleep 6
done
