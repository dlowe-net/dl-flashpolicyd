#!/bin/sh

if [ "x$GOPATH" == "x" ]; then
  echo GOPATH must be set for setup to work
  exit 1
fi

FILE=$1

if [ "x$FILE" == "x" ]; then
  echo Please specify the pathname of the policy file
  exit 1
fi

PROJECT=github.com/dlowe-net/dl-flashpolicyd
PROJDIR=$GOPATH/src/$PROJECT
go install $PROJECT

adduser --system --group \
  --home=/var/tmp \
  --no-create-home \
  --disabled-login \
  flashpolicyd

if [ -d /etc/logrotate.d ]; then
  install -o root -g root $PROJDIR/logrotate.conf /etc/logrotate.d/flashpolicyd
fi

install -o root -g root $GOPATH/bin/dl-flashpolicyd /usr/bin/flashpolicyd

install -d -o flashpolicyd -g flashpolicyd /var/log/flashpolicyd
chown -R flashpolicyd:flashpolicyd /var/log/flashpolicyd

install -m 755 $PROJDIR/init.sh /etc/init.d/flashpolicyd
install $PROJDIR/defaults.sh /etc/default/flashpolicyd

echo "FILE=\"$FILE\"" >>/etc/default/flashpolicyd

/usr/sbin/update-rc.d flashpolicyd defaults
