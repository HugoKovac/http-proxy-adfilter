#!/bin/sh

sleep 5

input_file="/lists/gambling.txt"

while IFS= read -r domain; do
	curl -x http://proxy:8888 -m 2 http://$domain
done < "$input_file"
