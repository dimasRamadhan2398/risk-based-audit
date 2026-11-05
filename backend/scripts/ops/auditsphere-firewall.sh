#!/bin/bash
# =============================================================================
# Block internet access to the internal ports Docker publishes on this VPS.
# Installed as /usr/local/sbin/auditsphere-firewall.sh, run at boot by
# auditsphere-firewall.service (oneshot, After=docker.service).
#
# Two different paths need two different rules:
#
#   IPv4: external traffic is DNATed in PREROUTING and goes through FORWARD,
#         so the rules live in DOCKER-USER, matched on the ORIGINAL host port
#         (ctorigdstport).
#   IPv6: Docker has IPv6 disabled here, so docker-proxy listens on [::]:<port>
#         on the HOST and relays to the container's IPv4 address. That traffic
#         goes through INPUT, never FORWARD, so DOCKER-USER does nothing for
#         it. (Found 2026-11-05: Kong admin :8009, the services and Prometheus
#         answered on 2001:df0:27b::529c:7684 although the v4 rules held.)
#
# nginx, the watchdog and the deploy scripts reach these ports on 127.0.0.1
# (lo), which none of these rules touch. Ports 22, 80 and 443 are never listed,
# so this cannot lock out SSH or the website.
#
# The long-term fix is binding the ports to 127.0.0.1 in compose; this script
# stays as defence in depth for anything published with a bare "port:port".
# Idempotent: each rule is checked before insert. Remove with --remove.
# =============================================================================
set -eu

# Interfaces that face the internet or a network we do not control.
# eth0 = public (v4 + v6), eth1 = provider private LAN (10.0.160.0/19).
# wg0 (10.0.0.0/24 WireGuard peer) is NOT listed: check nothing on the peer
# calls these ports before adding it.
# shellcheck disable=SC2206  # word-splitting the interface list is intended
IFACES=(${AUDITSPHERE_FW_IFACES:-eth0 eth1})

# Kong proxy/admin/TLS (prod+dev+tenant), backend services (prod+dev),
# monitoring (Prometheus, node exporter, Grafana, Loki), frontends, kafka-ui.
PORTS=(8080 8000 8001 8002 8003 8004 8009 8010 8019 8081 8084 8086 8090
       8100 8101 8102 8103 8104 8184 8443 8444 8445 8446
       9090 9100 9300 3100 3000 3001 3010 3020)

ACTION=insert
[ "${1:-}" = "--remove" ] && ACTION=remove

have_iface() { ip link show "$1" >/dev/null 2>&1; }

rule() { # $1 bin, $2 chain, rest = rule spec
  local bin="$1" chain="$2"; shift 2
  if [ "$ACTION" = remove ]; then
    while "$bin" -C "$chain" "$@" 2>/dev/null; do "$bin" -D "$chain" "$@"; done
  else
    "$bin" -C "$chain" "$@" 2>/dev/null || "$bin" -I "$chain" 1 "$@"
  fi
}

# IPv4 + IPv6 forwarded traffic (published ports reached via DNAT).
forward_rules() {
  local bin="$1" ifc p
  command -v "$bin" >/dev/null 2>&1 || return 0
  "$bin" -nL DOCKER-USER >/dev/null 2>&1 || "$bin" -N DOCKER-USER
  for ifc in "${IFACES[@]}"; do
    have_iface "$ifc" || continue
    for p in "${PORTS[@]}"; do
      rule "$bin" DOCKER-USER -i "$ifc" -p tcp -m conntrack --ctstate NEW \
        -m conntrack --ctorigdstport "$p" -j DROP
    done
  done
}

# IPv6 to docker-proxy on the host itself.
input_rules_v6() {
  local ifc p
  command -v ip6tables >/dev/null 2>&1 || return 0
  for ifc in "${IFACES[@]}"; do
    have_iface "$ifc" || continue
    for p in "${PORTS[@]}"; do
      rule ip6tables INPUT -i "$ifc" -p tcp --dport "$p" -j DROP
    done
  done
}

forward_rules iptables
forward_rules ip6tables
input_rules_v6
echo "auditsphere-firewall: ${ACTION} done on ${IFACES[*]} (${#PORTS[@]} ports)"
