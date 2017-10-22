#!/bin/sh

spawn-fcgi -s /var/run/fcgiwrap.socket /usr/bin/fcgiwrap
echo good
chmod 777 /var/run/fcgiwrap.socket

nginx -g daemon\ off\; &
/app/sfgit
