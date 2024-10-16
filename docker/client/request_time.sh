#! /bin/sh

PROXY_IP=$(getent hosts proxy | awk '{ print $1 }')
sleep 5

echo "First try without proxy"
curl -o /dev/null -w '@curl-format.txt' -s $1
echo "Second try without proxy"
curl -o /dev/null -w '@curl-format.txt' -s $1
iptables -t nat -A OUTPUT -p tcp --dport 80 -j DNAT --to-destination $PROXY_IP:7080
iptables -t nat -A OUTPUT -p tcp --dport 443 -j DNAT --to-destination $PROXY_IP:7443

echo "First try with proxy"
curl -o /dev/null -w '@curl-format.txt' -s $1
echo "Second try with proxy"
curl -o /dev/null -w '@curl-format.txt' -s $1
curl http://proxy:9000/add_sub_list --data category=base
echo "First try blocked by proxy"
curl -o /dev/null -w '@curl-format.txt' -s $1
curl http://proxy:9000/del_sub_list --data category=base