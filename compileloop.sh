#!/bin/sh
export PORT=8080
SOURCE=
for f in *.go; do
 # ../onthefly/*.go ../genericsite/*.go ../instapage/*.go ../simpleredis/*.go ../siteengines/*.go ../moskus/*.go ../webhandle/*.go; do
  SOURCE+="$f "
done
BIN=api
PIDFILE=/tmp/$BIN.pid
LOG=errors.err
M5=nop
SUMFILE=/tmp/$BIN.sumfile.txt
echo 'Starting compilation loop'
echo 'Reading pid'
if [ -e $PIDFILE ]; then
  echo 'Killing server'
  kill `cat $PIDFILE` > /dev/null
  rm $PIDFILE
fi
while true; do
  OLDM5=$M5
  md5sum $SOURCE > $SUMFILE
  M5=$(md5sum $SUMFILE)
  if [ "$OLDM5" != "$M5" ]; then
    echo 'Source changed'
    echo 'Reading pid'
    if [ -e $PIDFILE ]; then
      echo 'Killing server'
      kill `cat $PIDFILE` > /dev/null
      rm $PIDFILE
    fi
    clear
    date
    echo
    echo -n 'Recompiling API server...'
    [ -e $LOG ] && rm $LOG
    go build -o $BIN > $LOG
    if [ "$(wc -c $LOG | cut -d' ' -f1)" == '0' ]; then
      rm $LOG
    fi
    if [ -e $LOG ]; then
      echo
      cat $LOG
    else
      echo ok
    fi
    echo
    echo 'Backing up executable'
    if [ -e "/tmp/$BIN" ]; then
      rm "/tmp/$BIN"
    fi
    cp "./$BIN" "/tmp/$BIN"
    echo 'Starting server'
    PORT=$PORT ./$BIN &
    echo 'Writing pid'
    pgrep $BIN > $PIDFILE
  fi
  # Wait for the source to be changed
  inotifywait -q $SOURCE
  sleep 1
done
