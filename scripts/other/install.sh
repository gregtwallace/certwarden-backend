#/bin/sh

# install path and username
install_path="/opt/certwarden"
run_user="certwarden"


# Check for root
if [ "$(id -u)" -ne 0 ]; then echo "Please run as root." >&2; exit 1; fi

# move to script path
script_path=$(dirname $0)
cd "$script_path"

# create user to run app
useradd -r -s /bin/false "$run_user"

# copy all files to install path
mkdir "$install_path"
cp -R ../* "$install_path"

# permissions
./set_permissions.sh "$install_path" "$run_user"

# allow binding to low port numbers
setcap CAP_NET_BIND_SERVICE=+eip /opt/certwarden/certwarden

# install and start service
cp ./certwarden.service /etc/systemd/system/certwarden.service
systemctl daemon-reload
systemctl enable certwarden
systemctl start certwarden
