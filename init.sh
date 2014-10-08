#!/bin/sh
### BEGIN INIT INFO
# Provides:          dl-flashpolicyd
# Required-Start:    apache2 postgresql
# Required-Stop:     apache2 postgresql
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# X-Interactive:     true
# Short-Description: Start/stop dl-flashpolicyd flash policy server
### END INIT INFO

set -e

PATH=/sbin:/bin:/usr/sbin:/usr/bin
DAEMON=/usr/bin/flashpolicyd
NAME=flashpolicyd
DESC="flashpolicyd flash policy server"

PIDFILE=/var/run/$NAME.pid
SCRIPTNAME=/etc/init.d/$NAME

# Gracefully exit if the package has been removed.
test -x $DAEMON || exit 0

. /lib/lsb/init-functions

# Read config file if it is present.
if [ -r /etc/default/$NAME ]
then
    . /etc/default/$NAME
fi

FLASHPOLICYD_OPTS="-port=$PORT -user=$USER -file=$FILE -maxsize=$MAXSIZE -update=$UPDATE_EVERY -deadline=$CLIENT_DEADLINE"

ret=0
case "$1" in
  start)
    log_daemon_msg "Starting $DESC" "$NAME"
    if /usr/bin/daemon -n $NAME -r -D /  \
        -o /var/log/flashpolicyd/flashpolicyd.log -- $DAEMON $FLASHPOLICYD_OPTS
    then
        log_end_msg 0
    else
        ret=$?
        log_end_msg 1
    fi
    ;;
  stop)
    log_daemon_msg "Stopping $DESC" "$NAME"
    if /usr/bin/daemon -n $NAME --stop
    then
        log_end_msg 0
    else
        ret=$?
        log_end_msg 1
    fi
        rm -f $PIDFILE
    ;;
  reload|force-reload)
    log_action_begin_msg "Reloading $DESC configuration..."
    if /usr/bin/daemon -n $NAME --restart
    then
        log_action_end_msg 0
    else
        ret=$?
        log_action_end_msg 1
    fi
        ;;
  restart)
    $0 stop
    $0 start
    ret=$?
    ;;
  status)
    /usr/bin/daemon -n $NAME -u flashpolicyd:nobody --running
    ret=$?
    ;;

  *)
    echo "Usage: $SCRIPTNAME {start|stop|restart|reload|force-reload|status}" >&2
    exit 1
    ;;
esac

exit $ret
