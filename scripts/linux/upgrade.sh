#/bin/bash

# install path and username
lego_path=/opt/legocerthub
lego_user=legocerthub

# move to script path
script_path=$(dirname $0)
cd $script_path

# Check for root
if [ "$(id -u)" -ne 0 ]; then echo "Please run as root." >&2; exit 1; fi

# stop LeGo
systemctl stop legocerthub

# copy new files
rm -r $lego_path/frontend_build
cp -R ../* $lego_path

# permissions
chown $lego_user:$lego_user $lego_path -R
find $lego_path -type d -exec chmod 755 {} \;
find $lego_path -type f -exec chmod 640 {} \;
chmod 750 $lego_path/lego-linux-*
chmod 750 $lego_path/scripts/*.sh

# allow binding to low port numbers
case $(uname -m) in
    x86_64) setcap CAP_NET_BIND_SERVICE=+eip /opt/legocerthub/lego-linux-amd64 ;;
    arm)    setcap CAP_NET_BIND_SERVICE=+eip /opt/legocerthub/lego-linux-arm64 ;;
esac

# restart service
systemctl start legocerthub
