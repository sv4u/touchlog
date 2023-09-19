#!/bin/bash
HOST=cpanel.freehosting.com
USER=sasankvi
PASSWORD=`echo ${WEBSITE_ENC_KEY} | base64 --decude`

echo "Running 'make publish'"
make publish

echo "Moving into dist"
cd sit

echo "Transferring data"
ncftp -u $USER -p $PASSWORD $HOST <<EOF
cd public_html
mkdir log-suite
cd log-suite
rm -rf *
put -R .
bye
EOF

echo "Moving out of dist"
cd ..

echo "OK"

