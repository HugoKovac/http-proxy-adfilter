#!/bin/sh
PROXY_IP=$(getent hosts proxy | awk '{ print $1 }')
iptables -t nat -A OUTPUT -p tcp --dport 80 -j DNAT --to-destination $PROXY_IP:7080
iptables -t nat -A OUTPUT -p tcp --dport 443 -j DNAT --to-destination $PROXY_IP:7443

# Keep container running
tail -f /dev/null