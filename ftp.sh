#!/bin/bash
HOST=cpanel.freehosting.com
USER=sasankvi
PASSWORD=`echo ${WEBSITE_ENC_KEY} | base64 --decode`

echo "Moving into dist"
cd dist

echo "Transferring data"
ncftp -u $USER -p $PASSWORD $HOST <<EOF
cd domains/sasankvishnubhatla.net/public_html
mkdir log-suite
cd log-suite
mkdir touchlog
cd touchlog
rm -rf *
put -R .
bye
EOF

echo "Moving out of dist"
cd ..

echo "OK"

