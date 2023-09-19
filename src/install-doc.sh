#!/bin/bash

# pull ,touchlog from personsal site and store in temp file
temp_exec=$(mktemp)
touchlog=/usr/local/touchlog/bin/,touchlog
curl https://sasankvishnubhatla.net/log-suite/touchlog/,touchlog > $temp_exec
chmod a+x $temp_exec

# copy temp file into /usr/local/bin/,touchlog
echo Requiring sudo for copy into /usr/local/bin
sudo rm -rf /usr/local/touchlog/
sudo mkdir /usr/local/touchlog/
sudo mkdir /usr/local/touchlog/bin/
sudo cp $temp_exec $touchlog

# remove temp file
rm $temp_exec

# pull LICENSE from personsal site and store in temp file
temp_license=$(mktemp)
license=/usr/local/touchlog/LICENSE
curl https://sasankvishnubhatla.net/log-suite/touchlog/LICENSE > $temp_license

# copy temp file into /usr/local/,touchlog/
sudo cp $temp_license $license

# remove temp file
rm $temp_license

# pull README from personsal site and store in temp file
temp_readme=$(mktemp)
readme=/usr/local/touchlog/README.md
curl https://sasankvishnubhatla.net/log-suite/touchlog/README.md > $temp_readme

# copy temp file into /usr/local/touchlog/
sudo cp $temp_readme $readme

# remove temp file
rm $temp_readme

# pull ,touchlog.1 from personal site and store in temp file
temp_man=$(mktemp)
touchlog_man=,touchlog.1.gz
curl https://sasankvishnubhatla.net/log-suite/touchlog/,touchlog.1 > $temp_man

# gzip -cvf temp file > ,touchlog.1.gz
gzip -cvf $temp_man > $touchlog_man

# sudo cp ,touchlog.1.gz /usr/share/man/man1
sudo mv $touchlog_man /usr/share/man/man1

# sudo mandb
sudo mandb

# remove temp files
rm $temp_man

echo Please add /usr/local/touchlog/ to your path
echo Or run add the following line to your profile:
echo export PATH=\$PATH:/usr/local/touchlog/bin/
