#/bin/sh

# install path and username
lego_path="/opt/legocerthub"
lego_user="legocerthub"


# Check for root
if [ "$(id -u)" -ne 0 ]; then echo "Please run as root." >&2; exit 1; fi

# move to script path
script_path=$(dirname $0)
cd "$script_path"

# stop LeGo
systemctl stop legocerthub

# copy new files
rm -r "$lego_path"/frontend_build
cp -R ../* "$lego_path"

# permissions
./set_permissions.sh "$lego_path" "$lego_user"

# allow binding to low port numbers
case $(uname -m) in
    x86_64) setcap CAP_NET_BIND_SERVICE=+eip /opt/legocerthub/lego-linux-amd64 ;;
    arm)    setcap CAP_NET_BIND_SERVICE=+eip /opt/legocerthub/lego-linux-arm64 ;;
esac

# restart service
systemctl start legocerthub
