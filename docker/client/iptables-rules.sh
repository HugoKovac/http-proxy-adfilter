#!/bin/sh
PROXY_IP=$(getent hosts proxy | awk '{ print $1 }')
iptables -t nat -A OUTPUT -p tcp --dport 80 -j DNAT --to-destination $PROXY_IP:7080
iptables -t nat -A OUTPUT -p tcp --dport 443 -j DNAT --to-destination $PROXY_IP:7443

echo "$PROXY_IP proxy" >> /etc/hosts

sleep 40

curl https://www.google.com
curl http://proxy:9000/add_sub_list --data category=base

# Keep container running
tail -f /dev/null