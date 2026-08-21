#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

REFRESH_INTERVAL="${FIREWALL_REFRESH_INTERVAL:-15}"

ALLOWED_DOMAINS=(
  "api.anthropic.com"
  "console.anthropic.com"
  "claude.ai"
  "statsig.anthropic.com"
  "registry.npmjs.org"
  "deb.nodesource.com"
  "proxy.golang.org"
  "sum.golang.org"
  "storage.googleapis.com"
  "go.dev"
  "pypi.org"
  "files.pythonhosted.org"
)

resolve_domains() {
  for domain in "${ALLOWED_DOMAINS[@]}"; do
    local ips
    ips=$(dig +short A "$domain" 2>/dev/null || true)
    [ -z "$ips" ] && continue
    while read -r ip; do
      if [[ "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        ipset add allowed-domains "$ip" -exist 2>/dev/null || true
      fi
    done <<< "$ips"
  done
}

iptables -F
iptables -X 2>/dev/null || true
iptables -t nat -F
iptables -t nat -X 2>/dev/null || true
iptables -t mangle -F
iptables -t mangle -X 2>/dev/null || true
ipset destroy allowed-domains 2>/dev/null || true

iptables -A INPUT -i lo -j ACCEPT
iptables -A OUTPUT -o lo -j ACCEPT

iptables -A OUTPUT -p udp --dport 53 -j ACCEPT
iptables -A INPUT -p udp --sport 53 -j ACCEPT
iptables -A OUTPUT -p tcp --dport 53 -j ACCEPT
iptables -A INPUT -p tcp --sport 53 -j ACCEPT

iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

if command -v ip >/dev/null 2>&1; then
  HOST_IP=$(ip route | awk '/default/ {print $3; exit}')
  if [ -n "${HOST_IP:-}" ]; then
    HOST_NET=$(echo "$HOST_IP" | sed 's/\.[0-9]*$/.0\/24/')
    iptables -A INPUT -s "$HOST_NET" -j ACCEPT
    iptables -A OUTPUT -d "$HOST_NET" -j ACCEPT
  fi
fi

ipset create allowed-domains hash:net

echo "Fetching GitHub IP ranges..."
gh_ranges=$(curl -s https://api.github.com/meta || true)
if echo "$gh_ranges" | jq -e '.web and .api and .git' >/dev/null 2>&1; then
  echo "$gh_ranges" | jq -r '(.web + .api + .git)[]' | aggregate -q | while read -r cidr; do
    ipset add allowed-domains "$cidr" -exist 2>/dev/null || true
  done
else
  echo "WARNING: could not fetch GitHub ranges — go/pre-commit fetches from github may fail"
fi

resolve_domains

iptables -A OUTPUT -m set --match-set allowed-domains dst -j ACCEPT
iptables -A INPUT -m set --match-set allowed-domains src -j ACCEPT

iptables -P INPUT DROP
iptables -P OUTPUT DROP
iptables -P FORWARD DROP

echo "Firewall configured. Verifying..."
if curl -s --max-time 5 https://api.anthropic.com >/dev/null 2>&1; then
  echo "OK: api.anthropic.com reachable"
else
  echo "WARNING: api.anthropic.com NOT reachable"
fi
if curl -s --max-time 5 https://example.com >/dev/null 2>&1; then
  echo "ERROR: example.com reachable — allowlist is NOT restricting traffic; refusing to continue"
  exit 1
fi
echo "OK: example.com blocked as expected"

( while true; do sleep "$REFRESH_INTERVAL"; resolve_domains; done ) >/dev/null 2>&1 &
disown || true
echo "Started allowlist refresher (every ${REFRESH_INTERVAL}s) to track CDN IP rotation."
