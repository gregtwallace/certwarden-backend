#/bin/sh

# install path and username
lego_path="/opt/legocerthub"
lego_user="legocerthub"


# Check for root
if [ "$(id -u)" -ne 0 ]; then echo "Please run as root." >&2; exit 1; fi

# move to script path
script_path=$(dirname $0)
cd "$script_path"

# create user to run app
useradd -r -s /bin/false "$lego_user"

# copy all files to install path
mkdir "$lego_path"
cp -R ../* "$lego_path"

# make empty config
mkdir "$lego_path/data"
touch "$lego_path/data/config.yaml"

# permissions
./set_permissions.sh "$lego_path" "$lego_user"

# allow binding to low port numbers
case $(uname -m) in
    x86_64) setcap CAP_NET_BIND_SERVICE=+eip /opt/legocerthub/lego-linux-amd64 ;;
    arm)    setcap CAP_NET_BIND_SERVICE=+eip /opt/legocerthub/lego-linux-arm64 ;;
esac

# install and start service
cp ./legocerthub.service /etc/systemd/system/legocerthub.service
systemctl daemon-reload
systemctl enable legocerthub
systemctl start legocerthub
