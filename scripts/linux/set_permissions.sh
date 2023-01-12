#/bin/bash

chown legocerthub:legocerthub /opt/legocerthub -R

find /opt/legocerthub -type d -exec chmod 755 {} \;
find /opt/legocerthub -type f -exec chmod 640 {} \;

chmod 750 /opt/legocerthub/lego-amd64-linux

