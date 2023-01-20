#/bin/bash

# install path and username
lego_path=/opt/legocerthub
lego_user=legocerthub

# move to script path
script_path=$(dirname $0)
cd $script_path

# Check for root
if [ "$(id -u)" -ne 0 ]; then echo "Please run as root." >&2; exit 1; fi

echo "Enter the ip or hostname for LeGo. This is required to access LeGo"
echo "from a remote host. Leaving blank or setting to localhost will restrict"
echo "access to just localhost and will need to be updated in config.yaml later"
read -p 'Hostname [localhost]: ' hostname

hostname=${hostname:-localhost}

# create user to run app
useradd -r -s /bin/false $lego_user

# copy all files to install path
mkdir $lego_path
cp -R ../* $lego_path

# make config with hostname
printf "hostname: '$hostname'" > $lego_path/config.yaml

# permissions
chown $lego_user:$lego_user $lego_path -R
find $lego_path -type d -exec chmod 755 {} \;
find $lego_path -type f -exec chmod 640 {} \;
chmod 750 $lego_path/lego-linux-*
chmod 750 $lego_path/scripts/*.sh

# allow binding to low port numbers
setcap CAP_NET_BIND_SERVICE=+eip /opt/legocerthub/lego-linux-*

# install and start service
cp ./legocerthub.service /etc/systemd/system/legocerthub.service
systemctl daemon-reload
systemctl enable legocerthub
systemctl start legocerthub
