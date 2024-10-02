#!/bin/sh

set -e 

rm -f /root/activities.db

/root/migrate_glinet
cd /root && /root/filter_glinet
