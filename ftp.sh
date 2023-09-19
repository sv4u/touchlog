#!/bin/bash
HOST=cpanel.freehosting.com
USER=sasankvi
PASSWORD=`echo ${WEBSITE_ENC_KEY} | base64 --decode`

echo "Running 'make publish'"
make publish

echo "Moving into dist"
cd dist

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

