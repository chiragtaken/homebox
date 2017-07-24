#!/bin/sh

echo "Starting HomeBOX"
while [ true ]; do
    /bin/HboxServer -l debug > /var/log/hboxserver.log
    rm -f /var/log/hboxserver.log.earlier
    mv /var/log/hboxserver.log /var/log/hboxserver.log.earlier
    sleep 5
done &

while [ true ]; do
    sleep 5
done
