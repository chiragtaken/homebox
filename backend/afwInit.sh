#!/bin/sh

echo "Starting HomeDrop"
while [ true ]; do
    /bin/AfwServer -l debug > /var/log/afwserver.log
    rm -f /var/log/afwserver.log.earlier
    mv /var/log/afwserver.log /var/log/afwserver.log.earlier
    sleep 5
done &

while [ true ]; do
    sleep 5
done
