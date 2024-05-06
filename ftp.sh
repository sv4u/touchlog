#!/bin/bash
HOST=cpanel.freehosting.com
USER=sasankvi
PASSWORD=`echo ${WEBSITE_ENC_KEY} | base64 --decode`
PATH=domains/development.sasankvishnubhatla.net/public_html/log-suite/touchlog/

echo "Moving into dist"
cd dist

echo "Transferring data"
ncftpput -u $USER -p $PASSWORD $HOST $PATH *

echo "Moving out of dist"
cd ..

echo "OK"

