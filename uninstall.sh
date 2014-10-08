#!/bin/sh

service flashpolicyd stop
rm -f \
  /usr/bin/flashpolicyd \
  /etc/init.d/flashpolicyd \
  /etc/default/flashpolicyd \
  /etc/logrotate.d/flashpolicyd

/usr/sbin/update-rc.d flashpolicyd remove

chown -R root:root /var/log/flashpolicyd
deluser --system flashpolicyd
delgroup --only-if-empty flashpolicyd
