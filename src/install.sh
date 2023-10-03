#!/bin/bash

# pull ,touchlog from personsal site and store in temp file
temp_exec=$(mktemp)
touchlog=/usr/local/touchlog/bin/,touchlog
curl https://development.sasankvishnubhatla.net/log-suite/touchlog/,touchlog > $temp_exec
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
curl https://development.sasankvishnubhatla.net/log-suite/touchlog/LICENSE > $temp_license

# copy temp file into /usr/local/touchlog/
sudo cp $temp_license $license

# remove temp file
rm $temp_license

echo Please add /usr/local/touchlog/ to your path
echo Or run add the following line to your profile:
echo export PATH=\$PATH:/usr/local/touchlog/bin/
