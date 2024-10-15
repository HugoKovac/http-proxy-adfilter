#! /bin/sh

sleep 5

# curl -o /dev/null -w '@curl-format.txt' -s http://httpforever.com/
# curl -o /dev/null -x http://proxy:8888 -w '@curl-format.txt' -s http://httpforever.com/
curl -o /dev/null -x http://proxy:8888 -w '@curl-format.txt' -s http://stake.com
curl -o /dev/null -w '@curl-format.txt' -s -X POST http://proxy:8080/add_sub_list --data category=
curl -o /dev/null -x http://proxy:8888 -w '@curl-format.txt' -s http://stake.com
