#/bin/bash

lego_path=/opt/legocerthub
lego_user=legocerthub

chown $lego_user:$lego_user $lego_path -R

find $lego_path -type d -exec chmod 755 {} \;
find $lego_path -type f -exec chmod 640 {} \;

chmod 750 $lego_path/lego-linux-*
