#!/bin/sh

EXE=./airbrake-logger
OPTS="-listen 0.0.0.0:10000"
USER=wolf
LOG=/var/log/AirbrakeLogger/logger.log
chmod +x ${EXE}
exec chpst -u ${USER} ${EXE} ${OPTS} >> ${LOG} 2>&1
