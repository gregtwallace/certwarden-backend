#/bin/sh

# Default
default_path="/opt/legocerthub"
default_user="legocerthub"


# allow command line specified
lego_path="${1:-$default_path}"
lego_user="${2:-$default_user}"

# own entire folders / files
chown "$lego_user":"$lego_user" "$lego_path" -R

# set to user/owner read/edit only but allow directory browsing
find "$lego_path" -type d -exec chmod 755 {} \;
find "$lego_path" -type f -exec chmod 640 {} \;

# main executable
chmod 750 "$lego_path"/lego-linux-*

# allow execution of scripts
find "$lego_path"/scripts -type f -name "*.sh" -exec chmod 750 {} \;
