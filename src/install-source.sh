#!/bin/bash

# pull ,touchlog from personsal site and store in temp file
temp_exec=$(mktemp)
touchlog=/usr/local/touchlog/bin/,touchlog
curl https://sasankvishnubhatla.net/mood/touchlog/,touchlog > $temp_exec
chmod a+x $temp_exec

# copy temp file into /usr/local/touchlog/bin
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
curl https://sasankvishnubhatla.net/mood/touchlog/LICENSE > $temp_license

# copy temp file into /usr/local/touchlog/
sudo cp $temp_license $license

# remove temp file
rm $temp_license

# pull README from personsal site and store in temp file
temp_readme=$(mktemp)
readme=/usr/local/touchlog/README.md
curl https://sasankvishnubhatla.net/mood/touchlog/README.md > $temp_readme

# copy temp file into /usr/local/touchlog/
sudo cp $temp_readme $readme

# remove temp file
rm $temp_readme

# pull ,touchlog.c/.h and Makefile from personal site and store in temp files
temp_c=$(mktemp)
temp_h=$(mktemp)
temp_make=$(mktemp)
src_c=/usr/local/touchlog/src/,touchlog.c
src_h=/usr/local/touchlog/src/,touchlog.h
src_make=/usr/local/touchlog/src/Makefile
curl https://sasankvishnubhatla.net/mood/touchlog/,touchlog.c > $temp_c
curl https://sasankvishnubhatla.net/mood/touchlog/,touchlog.h > $temp_h
curl https://sasankvishnubhatla.net/mood/touchlog/Makefile > $temp_make

# copy temp fils to /usr/local/touchlog/src
sudo mkdir /usr/local/touchlog/src/
sudo cp $temp_c $src_c
sudo cp $temp_h $src_h
sudo cp $temp_make $src_make

# remove temp files
rm $temp_c
rm $temp_h
rm $temp_make

# pull ,touchlog.1 from personal site and store in temp file
temp_man=$(mktemp)
touchlog_man=,touchlog.1.gz
curl https://sasankvishnubhatla.net/mood/touchlog/,touchlog.1 > $temp_man

# gzip -cvf temp file > ,touchlog.1.gz
gzip -cvf $temp_man > $touchlog_man

# sudo cp ,touchlog.1.gz /usr/share/man/man1
sudo mv $touchlog_man /usr/share/man/man1

# sudo mandb
sudo mandb

# remove temp files
rm $temp_man

echo Please add /usr/local/touchlog/bin/ to your path
echo Or run add the following line to your profile:
echo export PATH=\$PATH:/usr/local/touchlog/bin
