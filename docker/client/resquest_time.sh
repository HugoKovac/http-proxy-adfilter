#! /bin/sh

sleep 5

curl -o /dev/null -w '@curl-format.txt' -s $1
curl -o /dev/null -x http://proxy:8888 -w '@curl-format.txt' -s $1
